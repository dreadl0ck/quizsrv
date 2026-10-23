#!/usr/bin/env python3
"""Build only from a verified source archive; assemble a fresh scratch context."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parent


def run(*args, **kwargs):
    subprocess.run(args, check=True, **kwargs)


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('manifest', type=Path)
    p.add_argument('output', type=Path)
    p.add_argument('image')
    a = p.parse_args()
    m = json.loads(a.manifest.read_text())
    archive = a.manifest.parent / (m['revision'] + '.tar.gz')
    assert hashlib.sha256(archive.read_bytes()).hexdigest() == m['sha256']
    assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT).decode().strip() == m['revision']
    run('git', 'diff', '--exit-code', 'HEAD', '--', '*.go', 'go.mod', 'go.sum', 'pages', 'files', 'courses', 'cmd', cwd=ROOT)
    a.output.mkdir(parents=True, exist_ok=False)
    source = a.output / 'source'
    source.mkdir()
    with tarfile.open(archive) as t:
        t.extractall(source, filter='data')
    source = next(source.iterdir())
    context = a.output / 'context'
    context.mkdir()
    env = dict(os.environ, GOOS='linux', GOARCH='amd64', CGO_ENABLED='0', GOPROXY='off', GOSUMDB='off')
    flags = f"-X main.SourceRevision={m['revision']} -X main.SourceSHA256={m['sha256']}"
    run('go', 'build', '-mod=vendor', '-trimpath', '-ldflags', flags, '-o', str(context / 'quizsrv'), '.', cwd=source, env=env)
    run('go', 'build', '-mod=vendor', '-trimpath', '-o', str(context / 'mdtopdf'), './cmd/mdtopdf', cwd=source, env=env)
    for directory in ['pages', 'files', 'courses']:
        (context / directory).mkdir()
        names = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', m['revision'], directory], cwd=ROOT).decode().splitlines()
        for name in names:
            path = Path(name)
            if any(p.startswith('._') or p == '.DS_Store' or 'SFPro' in p for p in path.parts): continue
            target = context / path
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(subprocess.check_output(['git', 'show', m['revision']+':'+name], cwd=ROOT))
    (context / 'source').mkdir()
    shutil.copy2(archive, context / 'source' / archive.name)
    for path in (source / 'vendor').rglob('*'):
        if path.is_file() and path.name.lower().startswith(('license', 'copying', 'notice', 'authors')):
            target = context / 'licenses' / path.relative_to(source / 'vendor')
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(path, target)
    for name in ['LICENSE', 'SOURCE-BUILD.md', 'sw.js', 'manifest.json']:
        if (source / name).exists(): shutil.copy2(source / name, context / name)
    shutil.copy2(ROOT / 'docker/Dockerfile', context / 'Dockerfile')
    (context / 'static').mkdir()
    run('docker', 'build', '--platform=linux/amd64', '--label', 'org.opencontainers.image.revision='+m['revision'], '--label', 'com.amsterdam-technologies.source.sha256='+m['sha256'], '-t', a.image, str(context))
    print(json.dumps({'image': a.image, 'revision': m['revision'], 'sourceSHA256': m['sha256']}))
