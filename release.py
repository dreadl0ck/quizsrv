#!/usr/bin/env python3
"""Verify published corresponding source and queue its build in Dokploy."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import re
import urllib.error
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parent


def run(*args, **kwargs):
    subprocess.run(args, check=True, **kwargs)


def api(endpoint, data=None, **query):
    base = os.environ['DOKPLOY_URL'].rstrip('/')
    if not base.startswith('https://'):
        raise ValueError('DOKPLOY_URL must use HTTPS')
    url = base + '/api/' + endpoint
    if query:
        url += '?' + urllib.parse.urlencode(query)
    request = urllib.request.Request(url, data=None if data is None else json.dumps(data).encode(),
        headers={'x-api-key': os.environ['DOKPLOY_API_KEY'], 'Content-Type': 'application/json'})
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            body = response.read()
            return json.loads(body) if body else None
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'Dokploy {endpoint}: HTTP {error.code}') from None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('manifest', type=Path)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    revision, digest = manifest['revision'], manifest['sha256']
    if not re.fullmatch(r'[0-9a-f]{40}', revision) or not re.fullmatch(r'[0-9a-f]{64}', digest):
        raise ValueError('Invalid source identity')
    if subprocess.check_output(['git', 'branch', '--show-current'], cwd=ROOT, text=True).strip() != 'main':
        raise ValueError('Release from main only')
    run('git', 'merge-base', '--is-ancestor', revision, 'origin/main', cwd=ROOT)
    remote = subprocess.check_output(['git', 'remote', 'get-url', 'origin'], cwd=ROOT, text=True).strip()
    match = re.search(r'/(learn|quizsrv)(?:\.git)?$', remote)
    if not match:
        raise ValueError('Unrecognized release repository')
    repo = match[1]
    identifier, domain = {'learn': ('ygoAInafQA5w-dGbKznyl', 'learn.dreadl0ck.net'),
        'quizsrv': ('d2mHSSYYUK20sV3FhInzK', 'sne.dreadl0ck.net')}[repo]
    url = f'https://{domain}/legal/source/{revision}.tar.gz'
    hashed = hashlib.sha256()
    with urllib.request.urlopen(url, timeout=120) as response:
        if response.url != url or response.status != 200:
            raise ValueError('Source must be anonymously available without redirect')
        while block := response.read(1024 * 1024):
            hashed.update(block)
    if hashed.hexdigest() != digest:
        raise ValueError('Public source hash mismatch')
    tag = 'dokploy-source-' + revision[:12]
    exists = subprocess.run(['git', 'show-ref', '--verify', '--quiet', 'refs/tags/' + tag], cwd=ROOT).returncode == 0
    if exists:
        if subprocess.check_output(['git', 'rev-parse', tag + '^{commit}'], cwd=ROOT, text=True).strip() != revision:
            raise ValueError('Release tag already names different source')
    else:
        run('git', 'tag', '-a', tag, revision, '-m', 'Dokploy source build', cwd=ROOT)
    run('git', 'push', 'origin', 'refs/tags/' + tag, cwd=ROOT)
    current = api('application.one', applicationId=identifier)
    if not any(d['host'] == domain for d in current['domains']):
        raise ValueError('Unexpected deployment domain')
    original_env, original_args = current.get('env'), current.get('buildArgs') or ''
    if original_args == '[REDACTED]':
        raise ValueError('Unredacted build arguments required')
    updated = original_args
    for key, value in {'SOURCE_REVISION': revision, 'SOURCE_SHA256': digest, 'VERSION': tag}.items():
        pattern = r'(?m)^' + key + r'=[^\r\n]*'
        if len(re.findall(pattern, updated)) > 1:
            raise ValueError('Duplicate source build argument')
        if re.search(pattern, updated):
            updated = re.sub(pattern, key + '=' + value, updated)
        else:
            updated += ('' if not updated or updated.endswith('\n') else '\n') + key + '=' + value + '\n'
    fresh = api('application.one', applicationId=identifier)
    if fresh.get('env') != original_env or (fresh.get('buildArgs') or '') != original_args:
        raise ValueError('Deployment configuration changed concurrently')
    api('application.update', {'applicationId': identifier, 'sourceType': 'github', 'owner': 'dreadl0ck',
        'repository': repo, 'githubId': os.environ.get('DOKPLOY_GITHUB_ID') or 'fCiPttXtULsq7nqK-Nf6E',
        'branch': tag, 'buildPath': '/', 'buildType': 'dockerfile', 'dockerfile': 'Dockerfile',
        'dockerContextPath': '.', 'buildArgs': updated, 'autoDeploy': False, 'triggerType': 'tag'})
    saved = api('application.one', applicationId=identifier)
    if saved.get('env') != original_env or saved['buildArgs'] != updated or saved['branch'] != tag:
        raise ValueError('Application environment/source configuration changed unexpectedly')
    api('application.deploy', {'applicationId': identifier, 'title': 'Dokploy source build ' + tag})
    print(json.dumps({'applicationId': identifier, 'repository': repo, 'sourceTag': tag,
        'sourceSHA256': digest, 'environmentPreserved': True, 'buildLocation': 'Dokploy estate server'}))


if __name__ == '__main__':
    main()
