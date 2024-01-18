package pattern

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestLint(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.Lint.Linters = map[string]*commands.Exec{
		"test": {
			Command: "test",
		},
	}

	os.MkdirAll("testdata/good", 0700)
	os.MkdirAll("testdata/bad", 0700)

	os.WriteFile("testdata/bad/main.jsonnet", []byte(`{
		run: [],
	}`), 0700)
	os.WriteFile("testdata/good/main.jsonnet", []byte(`{
		run: [
			{
				id: "a",
				change: "change",
				check: "check",
				remove: "remove"
			}
		]
	}`), 0700)

	r, err := Lint(ctx, c, "testdata/notreal", true)
	assert.HasErr(t, err, errs.ErrReceiver)
	assert.Equal(t, r, nil)

	d, _ := os.Getwd()

	r, err = Lint(ctx, c, "testdata/bad", true)
	assert.HasErr(t, err, nil)
	assert.Equal(t, r, types.Results{
		d + "/testdata/bad/main.jsonnet": {
			`diff have /main.jsonnet want /main.jsonnet
--- have /main.jsonnet
+++ want /main.jsonnet
@@ -1,3 +1,3 @@
 {
-		run: [],
-	}
\ No newline at end of file
+  run: [],
+}
`,
			"received empty commands",
		},
	})

	c.CLI.RunMockErrors([]error{
		ErrBuildEmpty,
	})
	c.CLI.RunMockOutputs([]string{
		"whoops",
	})

	r, err = Lint(ctx, c, "testdata/good", true)
	assert.HasErr(t, err, nil)
	assert.Equal(t, r, types.Results{
		d + "/testdata/good/main.jsonnet": {
			`diff have /main.jsonnet want /main.jsonnet
--- have /main.jsonnet
+++ want /main.jsonnet
@@ -1,10 +1,10 @@
 {
-		run: [
-			{
-				id: "a",
-				change: "change",
-				check: "check",
-				remove: "remove"
-			}
-		]
-	}
\ No newline at end of file
+  run: [
+    {
+      id: 'a',
+      change: 'change',
+      check: 'check',
+      remove: 'remove',
+    },
+  ],
+}
`,
			"linter test: whoops"},
	})

	os.RemoveAll("testdata")
}
