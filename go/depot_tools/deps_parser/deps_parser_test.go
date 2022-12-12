package deps_parser

import (
	"strings"
	"testing"

	"github.com/go-python/gpython/ast"
	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/deepequal/assertdeep"
)

const fakeDepsContent = `# This is a fake DEPS file.

# Extraneous declarations shouldn't trip us up.
unknown_thing = 'ok'

vars = {
  'variable_repo': 'https://my-host/var-repo.git',
  'variable_revision': 'var-revision',
  'format_repo': 'https://my-host/format-repo.git',
  'format_revision': 'format-revision',
  'dict_revision': 'dict-revision',
  'expr_prefix': 'expr/',
}

deps = {
  'simple/dep': 'https://my-host/simple-repo.git@simple-revision',
  'variable/dep': Var('variable_repo') + '@' + Var('variable_revision'),
  'format/dep' : '{format_repo}@{format_revision}',
  'dict/dep': {
    'url': 'https://my-host/dict-repo.git@' + Var('dict_revision'),
    'condition': 'always!!!1',
  },
  'cipd/deps': {
    'packages': [
      {
        'package': 'package1',
        'version': 'version1',
      },
      {
        'package': 'package2/${{platform}}',
        'version': 'version2',
      },
      {
        'package': 'package3/' + Var('host_os') + '-' + Var('host_cpu'),
        'version': 'pkg3-version',
      }
    ],
    'dep_type': 'cipd',
  },
  Var('expr_prefix')+'dep': 'https://my-host/expr-dep.git@version',
}
`

func TestParseDeps(t *testing.T) {

	// Verify that we parse the DEPS content successfully and get the
	// correct results for our toy example.
	deps, poss, err := parseDeps(fakeDepsContent)
	require.NoError(t, err)
	require.Equal(t, len(deps), len(poss))
	assertdeep.Equal(t, DepsEntries{
		"my-host/simple-repo": {
			Id:      "my-host/simple-repo",
			Version: "simple-revision",
			Path:    "simple/dep",
		},
		"my-host/var-repo": {
			Id:      "my-host/var-repo",
			Version: "var-revision",
			Path:    "variable/dep",
		},
		"my-host/format-repo": {
			Id:      "my-host/format-repo",
			Version: "format-revision",
			Path:    "format/dep",
		},
		"my-host/dict-repo": {
			Id:      "my-host/dict-repo",
			Version: "dict-revision",
			Path:    "dict/dep",
		},
		"package1": {
			Id:      "package1",
			Version: "version1",
			Path:    "cipd/deps",
		},
		"package2": {
			Id:      "package2",
			Version: "version2",
			Path:    "cipd/deps",
		},
		"package3/linux-x64": {
			Id:      "package3/linux-x64",
			Version: "pkg3-version",
			Path:    "cipd/deps",
		},
		"my-host/expr-dep": {
			Id:      "my-host/expr-dep",
			Version: "version",
			Path:    "expr/dep",
		},
	}, deps)

	// Positions should point to where the versions are defined.
	depsLines := strings.Split(fakeDepsContent, "\n")
	for idx, dep := range deps {
		pos := poss[idx]
		// Lineno starts at 1.
		lineIdx := pos.Lineno - 1
		str := depsLines[lineIdx][pos.ColOffset:]
		require.True(t, strings.Contains(str, dep.Version))
	}
}

func TestSetDep(t *testing.T) {

	before, beforePos, err := parseDeps(fakeDepsContent)
	require.NoError(t, err)
	beforeSplit := strings.Split(fakeDepsContent, "\n")

	// testSetDep runs SetDep and verifies that it performed the correct
	// modification.
	testSetDep := func(id, version string) {
		newDepsContent, err := SetDep(fakeDepsContent, id, version)
		require.NoError(t, err)
		after, afterPos, err := parseDeps(newDepsContent)
		require.NoError(t, err)
		var modifiedDepPos *ast.Pos
		for depId, dep := range after {
			expect := before[depId]
			if depId == NormalizeDep(id) {
				require.Equal(t, version, dep.Version)
				// Set the version back so we can use assertdeep.Equal.
				dep.Version = expect.Version
				modifiedDepPos = afterPos[depId]
			}
			assertdeep.Equal(t, expect, dep)
			assertdeep.Equal(t, beforePos[depId], afterPos[depId])
		}
		require.NotNil(t, modifiedDepPos, "Failed to find modified dep %q", id)
		modifiedLineIdx := modifiedDepPos.Lineno - 1 // Lineno starts at 1.

		// Ensure that we changed exactly one line.
		afterSplit := strings.Split(newDepsContent, "\n")
		require.Equal(t, len(beforeSplit), len(afterSplit))
		for idx, expectLine := range beforeSplit {
			actualLine := afterSplit[idx]
			if idx == modifiedLineIdx {
				require.NotEqual(t, expectLine, actualLine)
				require.True(t, strings.Contains(actualLine, version))
			} else {
				require.Equal(t, expectLine, actualLine)
			}
		}
	}

	// Run testSetDep for each dependency in fakeDepsContent.
	testSetDep("https://my-host/simple-repo.git", "newrev")
	testSetDep("my-host/simple-repo", "newrev") // Already normalized.
	testSetDep("https://my-host/var-repo.git", "newrev")
	testSetDep("https://my-host/dict-repo.git", "newrev")
	testSetDep("package1", "newrev")
	testSetDep("package2", "newrev")
}
