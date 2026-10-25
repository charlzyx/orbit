import importlib.util, io, pathlib, tarfile, tempfile, unittest

spec = importlib.util.spec_from_file_location('prepare_runtime', pathlib.Path(__file__).with_name('prepare-runtime.py'))
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)

class SafeArchiveTests(unittest.TestCase):
    def create(self, folder, entries):
        archive = folder / 'sample.tar.gz'
        with tarfile.open(archive, 'w:gz') as tf:
            for name, kind, content in entries:
                info = tarfile.TarInfo(name)
                if kind == 'file':
                    data = content.encode(); info.size = len(data); info.mode = 0o755
                    tf.addfile(info, io.BytesIO(data))
                elif kind == 'symlink':
                    info.type = tarfile.SYMTYPE; info.linkname = content; tf.addfile(info)
                elif kind == 'hardlink':
                    info.type = tarfile.LNKTYPE; info.linkname = content; tf.addfile(info)
        return archive
    def check_reject(self, entries):
        with tempfile.TemporaryDirectory() as tmp:
            p = pathlib.Path(tmp); archive = self.create(p, entries)
            with self.assertRaises(ValueError): m.safe_extract(archive, p / 'out', 'node-test')
    def test_file_and_internal_symlink(self):
        with tempfile.TemporaryDirectory() as tmp:
            p = pathlib.Path(tmp); a = self.create(p, [('node-test/lib/npm.js', 'file', 'ok'), ('node-test/bin/npm', 'symlink', '../lib/npm.js')])
            result = m.safe_extract(a, p / 'out', 'node-test')
            self.assertEqual(result['entries'], 2)
            self.assertEqual((p / 'out/node-test/bin/npm').read_text(), 'ok')
    def test_parent_traversal(self): self.check_reject([('node-test/../../bad', 'file', 'bad')])
    def test_absolute_path(self): self.check_reject([('/tmp/bad', 'file', 'bad')])
    def test_escape_symlink(self): self.check_reject([('node-test/bin/npm', 'symlink', '../../../bad')])
    def test_symlink_ancestor(self): self.check_reject([('node-test/lib', 'symlink', 'other'), ('node-test/lib/a', 'file', 'bad')])
    def test_hardlink(self): self.check_reject([('node-test/a', 'hardlink', 'node-test/b')])
    def test_duplicate(self): self.check_reject([('node-test/a', 'file', 'x'), ('node-test/a', 'file', 'y')])
    def test_member_budget(self):
        old = m.MAX_ENTRIES; m.MAX_ENTRIES = 1
        try: self.check_reject([('node-test/a', 'file', 'x'), ('node-test/b', 'file', 'y')])
        finally: m.MAX_ENTRIES = old
    def test_unpacked_budget(self):
        old = m.MAX_UNPACKED; m.MAX_UNPACKED = 1
        try: self.check_reject([('node-test/a', 'file', 'xy')])
        finally: m.MAX_UNPACKED = old

if __name__ == '__main__': unittest.main()
