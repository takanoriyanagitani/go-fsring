package fsring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvent(t *testing.T) {
	t.Parallel()

	var ITEST_FSRING_DIR string = os.Getenv("ITEST_FSRING_DIR")

	if len(ITEST_FSRING_DIR) < 1 {
		t.Skip("skipping tests using file system")
	}

	t.Run("WroteEventHandlerBuilderManaged", func(t *testing.T) {
		t.Parallel()

		var root string = filepath.Join(ITEST_FSRING_DIR, "WroteEventHandlerBuilderManaged")

		var chk NameChecker = NameCheckerNoCheck
		var mng string = "mng.txt"
		var hb WroteEventHandlerBuilder = WroteEventHandlerBuilderManaged(chk)(mng)

		var ngen NextName = NextNameDefault4

		var eh WroteEventHandler = hb(ngen)

		e := os.MkdirAll(root, 0755)
		mustNil(e)

		e = eh(WroteEvent{
			dir:       root,
			wroteName: "0000",
		})
		mustNil(e)
	})
}
