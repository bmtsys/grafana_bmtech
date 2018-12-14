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
			So(cmd.Result.SessionId, ShouldNotBeEmpty)
			So(cmd.Result.UserId, ShouldEqual, 1)
			So(cmd.Result.ClientIp, ShouldEqual, "192.168.10.11")
			So(cmd.Result.UserAgent, ShouldEqual, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
			So(cmd.Result.RefreshedAt, ShouldEqual, t.Unix())
			So(cmd.Result.CreatedAt, ShouldEqual, t.Unix())
			So(cmd.Result.UpdatedAt, ShouldEqual, t.Unix())

			Convey("Retrieve user session", func() {
				query := &m.GetUserSessionQuery{
					SessionID: cmd.Result.SessionId,
					UserID:    1,
				}

				err := GetUserSession(query)
				So(err, ShouldBeNil)
				So(query.Result, ShouldNotBeNil)
				So(query.Result.SessionId, ShouldNotBeEmpty)
				So(query.Result.UserId, ShouldEqual, 1)
				So(query.Result.ClientIp, ShouldEqual, "192.168.10.11")
				So(query.Result.UserAgent, ShouldEqual, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
				So(query.Result.RefreshedAt, ShouldEqual, t.Unix())
				So(query.Result.CreatedAt, ShouldEqual, t.Unix())
				So(query.Result.UpdatedAt, ShouldEqual, t.Unix())
			})
		})

		Reset(func() {
			now = time.Now
		})
	})
}
