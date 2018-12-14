package sqlstore

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	m "github.com/grafana/grafana/pkg/models"
)

func TestUserSession(t *testing.T) {
	Convey("Test user session", t, func() {
		InitTestDB(t)

		t := time.Date(2018, 12, 13, 13, 45, 0, 0, time.UTC)
		now = func() time.Time {
			return t
		}

		Convey("Create a user session", func() {
			cmd := &m.CreateUserSessionCommand{
				UserID:    1,
				ClientIP:  "192.168.10.11:1234",
				UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36",
			}

			err := CreateUserSession(cmd)
			So(err, ShouldBeNil)
			So(cmd.Result, ShouldNotBeNil)
			So(cmd.Result.Id, ShouldBeGreaterThan, 0)
			So(cmd.Result.UserId, ShouldEqual, 1)
			So(cmd.Result.AuthToken, ShouldNotBeEmpty)
			So(cmd.Result.PrevAuthToken, ShouldNotBeEmpty)
			So(cmd.Result.ClientIp, ShouldEqual, "192.168.10.11")
			So(cmd.Result.UserAgent, ShouldEqual, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
			So(cmd.Result.RotatedAt, ShouldEqual, t.Unix())
			So(cmd.Result.CreatedAt, ShouldEqual, t.Unix())
			So(cmd.Result.UpdatedAt, ShouldEqual, t.Unix())

			// Convey("Lookup user session by token", func() {
			// 	query := &m.LookupUserSessionByTokenQuery{
			// 		Token: cmd.Result.AuthToken,
			// 	}

			// 	err := LookupUserSessionByToken(query)
			// 	So(err, ShouldBeNil)
			// 	So(query.Result, ShouldNotBeNil)
			// 	So(query.Result.Id, ShouldBeGreaterThan, 0)
			// 	So(query.Result.UserId, ShouldEqual, 1)
			// 	So(query.Result.AuthToken, ShouldNotBeEmpty)
			// 	So(query.Result.PrevAuthToken, ShouldNotBeEmpty)
			// 	So(query.Result.ClientIp, ShouldEqual, "192.168.10.11")
			// 	So(query.Result.UserAgent, ShouldEqual, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
			// 	So(query.Result.RotatedAt, ShouldEqual, t.Unix())
			// 	So(query.Result.CreatedAt, ShouldEqual, t.Unix())
			// 	So(query.Result.UpdatedAt, ShouldEqual, t.Unix())
			// })
		})

		Reset(func() {
			now = time.Now
		})
	})
}

func TestParseIPAddress(t *testing.T) {
	Convey("Test parse ip address", t, func() {
		So(parseIPAddress("192.168.0.140:456"), ShouldEqual, "192.168.0.140")
		So(parseIPAddress("[::1:456]"), ShouldEqual, "127.0.0.1")
	})
}
