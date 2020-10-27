package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFeedbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Submit feedback to the team!",
		Run:   handleErrors(submitFeedback),
	}

	return cmd
}

func submitFeedback(cmd *cobra.Command, args []string) error {
	fmt.Println(`
Please email team@replicate.ai. We really appreciate your comments, good or bad!

				    ❤ Team Replicate\n`)
	return nil
}
