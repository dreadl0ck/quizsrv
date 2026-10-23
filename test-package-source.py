import importlib.util
import hashlib
import io
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('packager', Path(__file__).with_name('package-source.py'))
packager = importlib.util.module_from_spec(spec)
spec.loader.exec_module(packager)
publish_spec = importlib.util.spec_from_file_location('publisher', Path(__file__).with_name('publish-release.py'))
publisher = importlib.util.module_from_spec(publish_spec)
publish_spec.loader.exec_module(publisher)


class SourceSafety(unittest.TestCase):
    def test_excludes_course_and_generated_material(self):
        for name in ['courses/inr/deck.apkg', 'courses/inr/questions.yml', 'files/exam.pdf', 'files/anki/quiz.js', 'files/uploads/Inter.woff2', 'docker/mdtopdf', '.env', 'files/FontsFree-Net-SFProText-Regular.ttf']:
            self.assertFalse(packager.selected(name), name)

    def test_rejects_secret_and_nonregular_source(self):
        key = b'-----BEGIN ' + b'PRIVATE KEY-----\n' + b'A' * 64
        for name, mode, data in [('source.go', '100644', key), ('source.go', '120000', b'/etc/passwd'), ('cmd/.env', '100644', b'secret')]:
            with self.subTest(name=name), self.assertRaises(ValueError):
                packager.check(name, mode, data)

    def test_preserves_parser_delimiters(self):
        packager.check('vendor/parser.go', '100644', b'var marker = "-----BEGIN RSA ' + b'PRIVATE KEY-----\\n"')

    def test_source_publication_hash_gate(self):
        with patch.object(publisher.urllib.request, 'urlopen', return_value=io.BytesIO(b'wrong source')):
            with self.assertRaisesRegex(ValueError, 'publication blocked'):
                publisher.verify_public_source('https://source.example.test/archive', hashlib.sha256(b'expected source').hexdigest())

    def test_source_publication_needs_no_credentials(self):
        with patch.object(publisher.urllib.request, 'urlopen') as request:
            for url in ['http://source.example.test/archive', 'https://user:fixture@source.example.test/archive', 'https://source.example.test/archive?token=fixture']:
                with self.assertRaises(ValueError): publisher.verify_public_source(url, 'a'*64)
            request.assert_not_called()

    def test_historical_exports_require_matching_producer(self):
        with patch.object(packager, 'git', side_effect=[b'a'*40, b'different packager']):
            with self.assertRaisesRegex(ValueError, 'packager committed'):
                packager.package('HEAD', Path('unused-export'))


if __name__ == '__main__':
    unittest.main()
