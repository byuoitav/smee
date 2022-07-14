package commandcli

import (
	"net/http"

	avcli "github.com/byuoitav/smee/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Client) Screenshot(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no device provided; cancelling screenshot...")
		ctx.JSON(http.StatusBadRequest, "no device id")
	}

	cookie := ctx.Request.Header.Get("Cookie")
	token, err := parseForCookie("smee", cookie)
	if err != nil {
		c.log.Warn("authorization not found; cancelling screenshot...")
		ctx.JSON(http.StatusBadRequest, "authorization not found")
		return
	}

	netid, err := getUserFromJWT(token)
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling screenshot...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	resp, err := c.cli.Screenshot(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth)) // Todo: add auth?
	switch {
	case err != nil:
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.Unavailable:
				c.log.Warn("unable to get screenshot", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, "unable to get screenshot")
			case codes.Unauthenticated:
				c.log.Warn("unable to get screenshot", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, "unable to get screenshot")
			default:
				c.log.Warn("unable to get screenshot", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, "unable to get screenshot")
			}
		}
	case resp == nil:
		c.log.Warn("unable to get screenshot", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, "unable to get screenshot")
	}

	//use resp.GetPhoto() to get screenshot
	// Todo: find way to send screenshot

	ctx.JSON(http.StatusOK, id)
}
