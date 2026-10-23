#!/usr/bin/env python3
"""Publish an image only after anonymous matching source and embedded bytes pass."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tarfile
import urllib.parse
import urllib.request
import uuid


def verify_public_source(url, expected):
    parsed = urllib.parse.urlparse(url)
    if parsed.scheme != 'https' or parsed.username or parsed.password or parsed.query or parsed.fragment:
        raise ValueError('Source URL must be public HTTPS without credentials or query tokens')
    digest = hashlib.sha256()
    with urllib.request.urlopen(url, timeout=60) as response:
        while chunk := response.read(1024*1024): digest.update(chunk)
    if digest.hexdigest() != expected:
        raise ValueError('Public source hash mismatch; image publication blocked')


def verify_image(image, revision, expected):
    info = json.loads(subprocess.check_output(['docker', 'image', 'inspect', image]))[0]
    labels = info['Config']['Labels']
    if labels.get('org.opencontainers.image.revision') != revision or labels.get('com.amsterdam-technologies.source.sha256') != expected:
        raise ValueError('Image/source labels differ')
    container = 'source-publication-'+uuid.uuid4().hex[:12]
    subprocess.run(['docker', 'create', '--name', container, image], check=True, stdout=subprocess.DEVNULL)
    try:
        process = subprocess.Popen(['docker', 'cp', container+':/etc/quizsrv/source/'+revision+'.tar.gz', '-'], stdout=subprocess.PIPE)
        digest = hashlib.sha256()
        with tarfile.open(fileobj=process.stdout, mode='r|') as archive:
            for member in archive:
                if not member.isfile(): continue
                stream = archive.extractfile(member)
                while chunk := stream.read(1024*1024): digest.update(chunk)
        if process.wait() != 0 or digest.hexdigest() != expected:
            raise ValueError('Image contains mismatched source')
    finally:
        subprocess.run(['docker', 'rm', container], check=True, stdout=subprocess.DEVNULL)
    return info['Id']


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('manifest', type=Path)
    p.add_argument('image')
    p.add_argument('--source-url', required=True)
    p.add_argument('--check', action='store_true')
    args = p.parse_args()
    m = json.loads(args.manifest.read_text())
    if not re.fullmatch('[0-9a-f]{40}', m['revision']) or not re.fullmatch('[0-9a-f]{64}', m['sha256']):
        raise ValueError('Invalid source identity')
    verify_public_source(args.source_url, m['sha256'])
    identity = verify_image(args.image, m['revision'], m['sha256'])
    if json.loads(subprocess.check_output(['docker', 'image', 'inspect', args.image]))[0]['Id'] != identity:
        raise ValueError('Image changed during verification')
    if not args.check:
        subprocess.run(['docker', 'push', args.image], check=True)
    print(json.dumps({'image': args.image, 'imageId': identity, 'revision': m['revision'], 'sourceSHA256': m['sha256'], 'sourceFirstVerified': True, 'pushed': not args.check}))
