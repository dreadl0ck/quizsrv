import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('packager', Path(__file__).with_name('package-source.py'))
packager = importlib.util.module_from_spec(spec)
spec.loader.exec_module(packager)


class SourceSafety(unittest.TestCase):
    def test_excludes_course_and_generated_material(self):
        for name in ['courses/inr/deck.apkg', 'courses/inr/questions.yml', 'files/exam.pdf', 'docker/mdtopdf', '.env', 'files/FontsFree-Net-SFProText-Regular.ttf']:
            self.assertFalse(packager.selected(name), name)

    def test_rejects_secret_and_nonregular_source(self):
        key = b'-----BEGIN ' + b'PRIVATE KEY-----\n' + b'A' * 64
        for name, mode, data in [('source.go', '100644', key), ('source.go', '120000', b'/etc/passwd'), ('cmd/.env', '100644', b'secret')]:
            with self.subTest(name=name), self.assertRaises(ValueError):
                packager.check(name, mode, data)

    def test_preserves_parser_delimiters(self):
        packager.check('vendor/parser.go', '100644', b'var marker = "-----BEGIN RSA ' + b'PRIVATE KEY-----\\n"')


if __name__ == '__main__':
    unittest.main()
