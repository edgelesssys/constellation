import unittest
import os
import tempfile
import json
import parse
import compare

TEST_INPUTS = "./test-inputs"
RESULT_FNAME_AZURE = "constellation-azure.json"


class TestParse(unittest.TestCase):
    def test_parse(self):
        with tempfile.TemporaryDirectory() as tmpdirname:
            p = parse.BenchmarkParser(TEST_INPUTS, "azure", tmpdirname)
            p.parse()
            result_path = os.path.join(tmpdirname, RESULT_FNAME_AZURE)
            self.assertTrue(os.path.isfile(result_path))
            with open(result_path) as f:
                result = json.load(f)

            with open(os.path.join(TEST_INPUTS, RESULT_FNAME_AZURE)) as f:
                expected = json.load(f)
            self.assertEqual(result['fio'], expected['fio'])
            self.assertEqual(result['knb'], expected['knb'])


expected_comparison_result = '''# constellation-azure

<details>

- Commit of current benchmark: [N/A](https://github.com/edgelesssys/constellation/commit/N/A)
- Commit of previous benchmark: [N/A](https://github.com/edgelesssys/constellation/commit/N/A)

| Benchmark suite | Metric | Current | Previous | Ratio |
|-|-|-|-|-|
| read_iops | iops (IOPS) | 2165.847 | 2165.847 | 1.0 ⬆️ |
| write_iops | iops (IOPS) | 219.97105 | 219.97105 | 1.0 ⬆️ |
| read_bw | bw_kbytes (KiB/s) | 184151.0 | 184151.0 | 1.0 ⬆️ |
| write_bw | bw_kbytes (KiB/s) | 18604.0 | 18604.0 | 1.0 ⬆️ |
| pod2pod | tcp_bw_mbit (Mbit/s) | 943.0 | 943.0 | 1.0 ⬆️ |
| pod2pod | udp_bw_mbit (Mbit/s) | 595.0 | 595.0 | 1.0 ⬆️ |
| pod2svc | tcp_bw_mbit (Mbit/s) | 932.0 | 932.0 | 1.0 ⬆️ |
| pod2svc | udp_bw_mbit (Mbit/s) | 564.0 | 564.0 | 1.0 ⬆️ |

</details>'''


class TestCompare(unittest.TestCase):
    def test_compare(self):
        with tempfile.TemporaryDirectory() as tmpdirname:
            p = parse.BenchmarkParser(TEST_INPUTS, "azure", tmpdirname)
            p.parse()
            result_path = os.path.join(tmpdirname, RESULT_FNAME_AZURE)
            self.assertTrue(os.path.isfile(result_path))
            prev_path = os.path.join(TEST_INPUTS, RESULT_FNAME_AZURE)
            c = compare.BenchmarkComparer(prev_path, result_path)
            output = c.compare()
            self.assertEqual(output, expected_comparison_result)


if __name__ == '__main__':
    unittest.main()
