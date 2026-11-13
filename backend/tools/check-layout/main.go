// Command check-layout enforces backend folder and repository layout rules (wichat-go-layers.mdc).
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	backendRoot       string
	errs              []string
	allowedRepoProd   = regexp.MustCompile(`^(pg_repository|minio_repository|es_repository|redis_[a-z0-9_]+)\.go$`)
	kebabDir          = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	camelPkgDir       = map[string]bool{"test-support": true, "auth-redis": true, "auth-pb": true, "auth-http": true, "chat-http": true}
)

func main() {
	root, err := findBackendRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "check-layout: %v\n", err)
		os.Exit(1)
	}
	backendRoot = root

	for _, mod := range []string{"api-gateway", "ms-user", "ms-chat", "ms-notification", "wi-shared"} {
		modPath := filepath.Join(backendRoot, mod)
		if st, err := os.Stat(modPath); err != nil || !st.IsDir() {
			continue
		}
		checkGoMod(modPath)
		if mod != "wi-shared" {
			walkRepositoryDirs(filepath.Join(modPath, "internal"))
		} else {
			checkWiSharedRoot(modPath)
		}
		checkMsPkgLayout(modPath, mod)
		walkMisplacedTests(modPath)
		walkTestOneToOne(modPath)
	}
	checkOneToOneDomains(backendRoot)

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}
	fmt.Println("check-layout: ok")
}

func findBackendRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(cwd) == "backend" {
		return cwd, nil
	}
	if filepath.Base(cwd) == "scripts" && filepath.Base(filepath.Dir(cwd)) == "backend" {
		return filepath.Dir(cwd), nil
	}
	for d := cwd; d != filepath.Dir(d); d = filepath.Dir(d) {
		if filepath.Base(d) == "backend" {
			if _, err := os.Stat(filepath.Join(d, "Makefile")); err == nil {
				return d, nil
			}
		}
	}
	return "", fmt.Errorf("run from backend/ or repo root (could not find backend/)")
}

func checkGoMod(modPath string) {
	data, err := os.ReadFile(filepath.Join(modPath, "go.mod"))
	if err != nil {
		return
	}
	modName := filepath.Base(modPath)
	want := fmt.Sprintf("module github.com/wichat/wichat/backend/%s", modName)
	if !strings.Contains(string(data), want) {
		fail("%s: go.mod module path must be %q", modPath, want)
	}
}

func walkRepositoryDirs(internalRoot string) {
	_ = filepath.WalkDir(internalRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if filepath.Base(path) != "repository" {
			return nil
		}
		checkRepositoryDir(path)
		return nil
	})
}

func checkRepositoryDir(repoDir string) {
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if name == "mock" || name == "test" {
			if !e.IsDir() {
				fail("%s: %q must be a directory", repoDir, name)
			}
			continue
		}
		if e.IsDir() {
			fail("%s: unexpected subdir %q (repository root is whitelist files + mock/ + test/ only)", repoDir, name)
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			fail("%s: production repository must not contain %q (use repository/test/)", repoDir, name)
			continue
		}
		if !allowedRepoProd.MatchString(name) {
			fail("%s: disallowed repository file %q", repoDir, name)
		}
	}

	mockDir := filepath.Join(repoDir, "mock")
	pgFile := filepath.Join(repoDir, "pg_repository.go")
	if st, err := os.Stat(pgFile); err == nil && !st.IsDir() {
		if _, err := os.Stat(mockDir); err != nil {
			fail("%s: missing mock/pg_repository_mock.go (required when pg_repository.go exists)", repoDir)
			return
		}
		mockFile := filepath.Join(mockDir, "pg_repository_mock.go")
		if _, err := os.Stat(mockFile); err != nil {
			fail("%s: missing mock/pg_repository_mock.go", repoDir)
		}
	}
}

func checkWiSharedRoot(modPath string) {
	skip := map[string]bool{"infra": true, "models": true, "microservices": true}
	entries, err := os.ReadDir(modPath)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		name := e.Name()
		if skip[name] {
			continue
		}
		if strings.Contains(name, "_") {
			fail("%s/%s: wi-shared top-level dirs use kebab-case, not snake_case", modPath, name)
		}
		if strings.Contains(name, "chat") && strings.Contains(name, "event") && !strings.Contains(name, "-") {
			fail("%s/%s: rename to chat-events (kebab-case dir, chatEvents package)", modPath, name)
		}
		if !kebabDir.MatchString(name) && strings.Contains(name, "-") {
			fail("%s/%s: invalid wi-shared dir name %q", modPath, name, name)
		}
	}
}

func checkMsPkgLayout(modPath, modName string) {
	if modName == "wi-shared" {
		return
	}
	pkgRoot := filepath.Join(modPath, "pkg")
	if st, err := os.Stat(pkgRoot); err != nil || !st.IsDir() {
		return
	}
	entries, err := os.ReadDir(pkgRoot)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "constants" || kebabDir.MatchString(name) {
			if strings.Contains(name, "_") {
				fail("%s/pkg/%s: use kebab-case dir names, not snake_case", modPath, name)
			}
			continue
		}
		if camelPkgDir[name] {
			continue
		}
		if !kebabDir.MatchString(name) {
			fail("%s/pkg/%s: multi-word pkg dirs must be kebab-case (e.g. auth-pb, test-support)", modPath, name)
		}
	}
}

func walkTestOneToOne(modPath string) {
	_ = filepath.WalkDir(modPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || filepath.Base(path) != "test" {
			return nil
		}
		if strings.Contains(path, string(os.PathSeparator)+"tools"+string(os.PathSeparator)) {
			return nil
		}
		pkgDir := filepath.Dir(path)
		checkTestOneToOne(pkgDir, path)
		return nil
	})
}

func checkTestOneToOne(pkgDir, testDir string) {
	prods := listProdGoFiles(pkgDir)
	if len(prods) == 0 {
		return
	}

	expected := make(map[string]string, len(prods))
	for _, p := range prods {
		expected[strings.TrimSuffix(p, ".go")+"_test.go"] = p
	}

	entries, err := os.ReadDir(testDir)
	if err != nil {
		return
	}

	seen := make(map[string]bool, len(expected))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			fail("%s: test/ must not contain subdir %q", testDir, name)
			continue
		}
		if !strings.HasSuffix(name, "_test.go") {
			if strings.HasSuffix(name, ".go") {
				fail("%s: non-test file %q (only *_test.go allowed in test/)", testDir, name)
			}
			continue
		}
		prod, ok := expected[name]
		if !ok {
			hint := mergeHint(prods, name)
			fail("%s: extra test file %q — merge into test/%s (1:1 with %s)", testDir, name, hint, strings.Join(prods, ", "))
			continue
		}
		seen[name] = true
		_ = prod
	}

	for want, prod := range expected {
		if !seen[want] {
			fail("%s: missing %s for production %s", testDir, want, prod)
		}
	}
}

func listProdGoFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_exported_test.go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		out = append(out, name)
	}
	return out
}

func mergeHint(prods []string, extraTest string) string {
	if len(prods) == 1 {
		return strings.TrimSuffix(prods[0], ".go") + "_test.go"
	}
	extraStem := strings.TrimSuffix(extraTest, "_test.go")
	for _, p := range prods {
		prodStem := strings.TrimSuffix(p, ".go")
		if strings.Contains(extraStem, prodStem) || strings.Contains(extraTest, prodStem) {
			return prodStem + "_test.go"
		}
	}
	return strings.TrimSuffix(prods[0], ".go") + "_test.go"
}

func walkMisplacedTests(modPath string) {
	_ = filepath.WalkDir(modPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		if filepath.Base(path) == "main_test.go" {
			return nil
		}
		if strings.Contains(path, string(os.PathSeparator)+"tools"+string(os.PathSeparator)) {
			return nil
		}
		dir := filepath.Dir(path)
		if filepath.Base(dir) == "test" {
			return nil
		}
		fail("%s: move tests into a test/ subdir (%s)", path, dir)
		return nil
	})
}

func fail(format string, args ...any) {
	errs = append(errs, "check-layout: "+fmt.Sprintf(format, args...))
}

func checkOneToOneDomains(root string) {
	msInternal := filepath.Join(root, "ms-user", "internal")
	grpcClientsDir := filepath.Join(root, "wi-shared", "microservices", "services", "ms-user", "grpc")
	protoDir := filepath.Join(root, "wi-shared", "models", "proto", "user", "v1")

	entries, err := os.ReadDir(msInternal)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		domain := e.Name()
		grpcServer := filepath.Join(msInternal, domain, "delivery", "grpc", "server.go")
		if _, err := os.Stat(grpcServer); err != nil {
			continue
		}
		clientFile := filepath.Join(grpcClientsDir, domain, "client.go")
		if _, err := os.Stat(clientFile); err != nil {
			fail("ms-user/internal/%s has gRPC server.go, missing wi-shared client %s", domain, clientFile)
		}
		protoFile := filepath.Join(protoDir, strings.ReplaceAll(domain, "-", "_")+".proto")
		if _, err := os.Stat(protoFile); err != nil {
			fail("ms-user/internal/%s has gRPC server.go, missing proto %s", domain, protoFile)
		}
	}

	clientEntries, err := os.ReadDir(grpcClientsDir)
	if err != nil {
		return
	}
	for _, e := range clientEntries {
		if !e.IsDir() {
			continue
		}
		domain := e.Name()
		grpcServer := filepath.Join(msInternal, domain, "delivery", "grpc", "server.go")
		if _, err := os.Stat(grpcServer); err != nil {
			fail("wi-shared ms-user/grpc/%s exists, missing ms-user/internal/%s/delivery/grpc/server.go", domain, domain)
		}
	}
}
