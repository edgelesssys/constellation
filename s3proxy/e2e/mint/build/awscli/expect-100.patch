diff --git a/botocore/handlers.py b/botocore/handlers.py
index 4b1d6445e..c8dd79c3e 100644
--- a/botocore/handlers.py
+++ b/botocore/handlers.py
@@ -333,9 +333,11 @@ def add_expect_header(model, params, **kwargs):
         return
     if 'body' in params:
         body = params['body']
-        if hasattr(body, 'read'):
+        size = getattr(body, "_size", None)
+        if size:
             # Any file like object will use an expect 100-continue
-            # header regardless of size.
+            # except when size equals zero.
+            # https://tools.ietf.org/html/rfc7231#section-5.1.1
             logger.debug("Adding expect 100 continue header to request.")
             params['headers']['Expect'] = '100-continue'
 
diff --git a/tests/unit/test_awsrequest.py b/tests/unit/test_awsrequest.py
index 22bd9a746..44cd05f61 100644
--- a/tests/unit/test_awsrequest.py
+++ b/tests/unit/test_awsrequest.py
@@ -504,6 +504,14 @@ class TestAWSHTTPConnection(unittest.TestCase):
         response = conn.getresponse()
         self.assertEqual(response.status, 200)
 
+    def test_no_expect_header_set_no_body(self):
+        s = FakeSocket(b'HTTP/1.1 200 OK\r\n')
+        conn = AWSHTTPConnection('s3.amazonaws.com', 443)
+        conn.sock = s
+        conn.request('PUT', '/bucket/foo', b'')
+        response = conn.getresponse()
+        self.assertEqual(response.status, 200)
+
     def test_tunnel_readline_none_bugfix(self):
         # Tests whether ``_tunnel`` function is able to work around the
         # py26 bug of avoiding infinite while loop if nothing is returned.
