package ftp

import (
	"context"
	"strconv"
)

/*
   RESTART (REST)

      The argument field represents the server marker at which
      file transfer is to be restarted.  This command does not
      cause file transfer but skips over the file to the specified
      data checkpoint.  This command shall be immediately followed
      by the appropriate FTP service command which shall cause
      file transfer to resume.
*/

type commandREST struct{}

func (c commandREST) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandREST) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	position, err := strconv.Atoi(params[0])
	if err != nil {
		return s.ReplyStatus(StatusSyntaxError)
	}

	s.restartPosition = position

	return s.ReplyStatus(StatusPendingMoreInfo)
}

func init() {
	commandMap["REST"] = &commandREST{}
}
