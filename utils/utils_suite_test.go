package utils_test

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/operating"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpbackup/testutils"
	"github.com/greenplum-db/gpbackup/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var (
	connectionPool *dbconn.DBConn
	mock           sqlmock.Sqlmock
	stdout         *gbytes.Buffer
	stderr         *gbytes.Buffer
	logfile        *gbytes.Buffer
	buffer         *gbytes.Buffer
	toc            *utils.TOC
	backupfile     *utils.FileWithByteCount
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "utils tests")
}

var _ = BeforeSuite(func() {
	connectionPool, mock, stdout, stderr, logfile = testutils.SetupTestEnvironment()
})

var _ = BeforeEach(func() {
	connectionPool, mock, stdout, stderr, logfile = testhelper.SetupTestEnvironment()
	operating.System = operating.InitializeSystemFunctions()
	buffer = gbytes.NewBuffer()
})
