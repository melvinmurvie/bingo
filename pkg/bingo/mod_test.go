// Copyright (c) Bartłomiej Płotka @bwplotka
// Licensed under the Apache License 2.0.

package bingo

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/efficientgo/core/testutil"
	"github.com/melvinmurvie/bingo/pkg/runner"
	"github.com/melvinmurvie/bingo/pkg/version"
	"golang.org/x/mod/module"
)

func goVersion(r *runner.Runner) string {
	// Starting from Go 1.21, `go mod init` adds complete semver to modfile.
	// Thus, we return <major>.<minor> for < 1.21, and full semver otherwise.
	if r.GoVersion().Compare(version.Go121) == -1 {
		return fmt.Sprintf("%v.%v", r.GoVersion().Major(), r.GoVersion().Minor())
	}

	return r.GoVersion().String()
}

func TestCreateFromExistingOrNew(t *testing.T) {
	logger := log.New(os.Stderr, "", 0)
	r, err := runner.NewRunner(context.TODO(), logger, false, "go")
	testutil.Ok(t, err)

	t.Run("create new and close should create empty mod file with basic autogenerated meta", func(t *testing.T) {
		f, err := CreateFromExistingOrNew(context.TODO(), r, log.New(os.Stderr, "", 0), "non_existing.mod", "test.mod")
		testutil.Ok(t, err)
		testutil.Ok(t, f.Close())

		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s
`, goVersion(r)), "test.mod")
	})
	t.Run("create new and close should work and produce same output", func(t *testing.T) {
		f, err := CreateFromExistingOrNew(context.TODO(), r, log.New(os.Stderr, "", 0), "test.mod", "test2.mod")
		testutil.Ok(t, err)
		testutil.Ok(t, f.Close())
		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s
`, goVersion(r)), "test.mod")
		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s
`, goVersion(r)), "test2.mod")
	})
	t.Run("create new and set direct require should work", func(t *testing.T) {
		f, err := CreateFromExistingOrNew(context.TODO(), r, log.New(os.Stderr, "", 0), "", "test3.mod")
		testutil.Ok(t, err)
		testutil.Ok(t, f.SetDirectRequire(Package{Module: module.Version{Path: "github.com/yolo/best/v100", Version: "v100.0.0"}, RelPath: "thebest"}))
		testutil.Equals(t, Package{Module: module.Version{Path: "github.com/yolo/best/v100", Version: "v100.0.0"}, RelPath: "thebest"}, *f.DirectPackage())
		testutil.Ok(t, f.Close())
		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s

require github.com/yolo/best/v100 v100.0.0 // thebest
`, goVersion(r)), "test3.mod")
	})
	t.Run("create new and set direct require2 should work", func(t *testing.T) {
		f, err := CreateFromExistingOrNew(context.TODO(), r, log.New(os.Stderr, "", 0), "", "test4.mod")
		testutil.Ok(t, err)
		testutil.Ok(t, f.SetDirectRequire(Package{Module: module.Version{Path: "github.com/yolo/best/v100", Version: "v100.0.0"}}))
		testutil.Equals(t, Package{Module: module.Version{Path: "github.com/yolo/best/v100", Version: "v100.0.0"}}, *f.DirectPackage())
		testutil.Ok(t, f.Close())
		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s

require github.com/yolo/best/v100 v100.0.0
`, goVersion(r)), "test4.mod")
	})
	t.Run("copy and set direct require to something else", func(t *testing.T) {
		f, err := CreateFromExistingOrNew(context.TODO(), r, log.New(os.Stderr, "", 0), "test3.mod", "test5.mod")
		testutil.Ok(t, err)
		testutil.Equals(t, Package{Module: module.Version{Path: "github.com/yolo/best/v100", Version: "v100.0.0"}, RelPath: "thebest"}, *f.DirectPackage())
		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s

require github.com/yolo/best/v100 v100.0.0 // thebest
`, goVersion(r)), "test5.mod")

		testutil.Ok(t, f.SetDirectRequire(Package{Module: module.Version{Path: "github.com/yolo/not-best", Version: "v1"}}))
		testutil.Equals(t, Package{Module: module.Version{Path: "github.com/yolo/not-best", Version: "v1"}}, *f.DirectPackage())
		testutil.Ok(t, f.Close())
		expectContent(t, fmt.Sprintf(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go %s

require github.com/yolo/not-best v1
`, goVersion(r)), "test5.mod")
	})
}

func expectContent(t *testing.T, expected string, file string) {
	t.Helper()

	b, err := os.ReadFile(file)
	testutil.Ok(t, err)
	testutil.Equals(t, expected, string(b))
}

func TestModFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("without auto fetch directives", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.mod")
		testutil.Ok(t, os.WriteFile(testFile, []byte(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go 1.14

// bingo:no_directive_fetch

replace (
	// Ridiculous but Prometheus v2.4.3 did not have Go modules
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v5.0.0-beta.0.20161028183111-bd73d950fa44+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v9.9.0+incompatible
	github.com/cockroachdb/cmux => github.com/cockroachdb/cmux v0.0.0-20170110192607-30d10be49292
	github.com/cockroachdb/cockroach => github.com/cockroachdb/cockroach v0.0.0-20170608034007-84bc9597164f
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.2.3-0.20180520015035-48a0ecefe2e4
	github.com/miekg/dns => github.com/miekg/dns v1.0.4
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.0-pre1.0.20180607123607-faf4ec335fe0
	github.com/prometheus/common => github.com/prometheus/common v0.0.0-20180518154759-7600349dcfe1
	github.com/prometheus/tsdb => github.com/prometheus/tsdb v0.0.0-20180921053122-9c8ca47399a7
	k8s.io/api => k8s.io/api v0.0.0-20180628040859-072894a440bd
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d
	k8s.io/client-go => k8s.io/client-go v8.0.0+incompatible
	k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0
)

require github.com/prometheus/prometheus v2.4.3+incompatible // cmd/prometheus
`), os.ModePerm))

		mf, err := OpenModFile(testFile)
		testutil.Ok(t, err)

		testutil.Equals(t, true, mf.IsDirectivesAutoFetchDisabled())
		testutil.Equals(t, Package{Module: module.Version{Path: "github.com/prometheus/prometheus", Version: "v2.4.3+incompatible"}, RelPath: "cmd/prometheus"}, *mf.DirectPackage())
	})

	t.Run("with auto fetch directives", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.mod")
		testutil.Ok(t, os.WriteFile(testFile, []byte(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go 1.14

replace (
	// Ridiculous but Prometheus v2.4.3 did not have Go modules
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v5.0.0-beta.0.20161028183111-bd73d950fa44+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v9.9.0+incompatible
	github.com/cockroachdb/cmux => github.com/cockroachdb/cmux v0.0.0-20170110192607-30d10be49292
	github.com/cockroachdb/cockroach => github.com/cockroachdb/cockroach v0.0.0-20170608034007-84bc9597164f
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.2.3-0.20180520015035-48a0ecefe2e4
	github.com/miekg/dns => github.com/miekg/dns v1.0.4
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.0-pre1.0.20180607123607-faf4ec335fe0
	github.com/prometheus/common => github.com/prometheus/common v0.0.0-20180518154759-7600349dcfe1
	github.com/prometheus/tsdb => github.com/prometheus/tsdb v0.0.0-20180921053122-9c8ca47399a7
	k8s.io/api => k8s.io/api v0.0.0-20180628040859-072894a440bd
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d
	k8s.io/client-go => k8s.io/client-go v8.0.0+incompatible
	k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0
)

require github.com/prometheus/prometheus v2.4.3+incompatible // cmd/prometheus
`), os.ModePerm))

		mf, err := OpenModFile(testFile)
		testutil.Ok(t, err)

		testutil.Equals(t, false, mf.IsDirectivesAutoFetchDisabled())
		testutil.Equals(t, Package{Module: module.Version{Path: "github.com/prometheus/prometheus", Version: "v2.4.3+incompatible"}, RelPath: "cmd/prometheus"}, *mf.DirectPackage())
	})

	t.Run("with build attributes1", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.mod")
		testutil.Ok(t, os.WriteFile(testFile, []byte(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go 1.14

require github.com/prometheus/prometheus v2.4.3+incompatible // cmd/prometheus -tags=yolo,linux
`), os.ModePerm))

		mf, err := OpenModFile(testFile)
		testutil.Ok(t, err)

		testutil.Equals(t, false, mf.IsDirectivesAutoFetchDisabled())
		testutil.Equals(t, Package{
			Module:     module.Version{Path: "github.com/prometheus/prometheus", Version: "v2.4.3+incompatible"},
			RelPath:    "cmd/prometheus",
			BuildFlags: []string{"-tags=yolo,linux"},
		}, *mf.DirectPackage())
	})

	t.Run("with build attributes2", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.mod")
		testutil.Ok(t, os.WriteFile(testFile, []byte(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go 1.14

require github.com/prometheus/prometheus v2.4.3+incompatible // cmd/prometheus CGO_ENABLED=1 GOWASM=somefeature
`), os.ModePerm))

		mf, err := OpenModFile(testFile)
		testutil.Ok(t, err)

		testutil.Equals(t, false, mf.IsDirectivesAutoFetchDisabled())
		testutil.Equals(t, Package{
			Module:    module.Version{Path: "github.com/prometheus/prometheus", Version: "v2.4.3+incompatible"},
			RelPath:   "cmd/prometheus",
			BuildEnvs: []string{"CGO_ENABLED=1", "GOWASM=somefeature"},
		}, *mf.DirectPackage())
	})

	t.Run("with build attributes3", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.mod")
		testutil.Ok(t, os.WriteFile(testFile, []byte(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go 1.14

require github.com/prometheus/prometheus v2.4.3+incompatible // cmd/prometheus CGO_ENABLED=1 GOWASM=somefeature -tags=yolo,linux
`), os.ModePerm))

		mf, err := OpenModFile(testFile)
		testutil.Ok(t, err)

		testutil.Equals(t, false, mf.IsDirectivesAutoFetchDisabled())
		testutil.Equals(t, Package{
			Module:     module.Version{Path: "github.com/prometheus/prometheus", Version: "v2.4.3+incompatible"},
			RelPath:    "cmd/prometheus",
			BuildEnvs:  []string{"CGO_ENABLED=1", "GOWASM=somefeature"},
			BuildFlags: []string{"-tags=yolo,linux"},
		}, *mf.DirectPackage())
	})
	t.Run("with build attributes without relpath", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.mod")
		testutil.Ok(t, os.WriteFile(testFile, []byte(`module _ // Auto generated by https://github.com/bwplotka/bingo. DO NOT EDIT

go 1.14

require github.com/prometheus/prometheus v2.4.3+incompatible // CGO_ENABLED=1 GOWASM=somefeature -tags=yolo,linux
`), os.ModePerm))

		mf, err := OpenModFile(testFile)
		testutil.Ok(t, err)

		testutil.Equals(t, false, mf.IsDirectivesAutoFetchDisabled())
		testutil.Equals(t, Package{
			Module:     module.Version{Path: "github.com/prometheus/prometheus", Version: "v2.4.3+incompatible"},
			BuildEnvs:  []string{"CGO_ENABLED=1", "GOWASM=somefeature"},
			BuildFlags: []string{"-tags=yolo,linux"},
		}, *mf.DirectPackage())
	})
}
