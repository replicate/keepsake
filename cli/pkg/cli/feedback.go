package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFeedbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Submit feedback to the team!",
		RunE:  submitFeedback,
	}

	return cmd
}

func submitFeedback(cmd *cobra.Command, args []string) error {
	fmt.Println(`
The feedback function hasn't yet been implemented in the CLI.
In the meantime, Please visit beta.replicate.ai and click the
Feedback button in the bottom right corner.

We really appreciate your feedback, good or bad!

				    ❤ Team Replicate\n`)
	return nil
}
