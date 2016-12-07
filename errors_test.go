package errors_test

import (
	"fmt"
	"time"

	"github.com/atdiar/errors"
)

func Example() {
	v := errors.New("Something happened.")
	v = v.AddInfo("date", time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Format(time.UnixDate))
	v = v.AddInfo(errors.PrintLine())
	v = v.AddInfo(errors.PrintFunc())

	var vIface error

	vIface = v

	fmt.Print(vIface)
	// Output:
	//{
	//  "ErrorInfo": {
	//   "date": "Tue Nov 10 23:00:00 UTC 2009",
	//   "fn": "github.com/atdiar/errors_test.Example",
	//   "line": 13
	//  },
	//  "ErrorCause": "Something happened."
	//}
}
