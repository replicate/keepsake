package analytics

import (
	"fmt"

	"replicate.ai/cli/pkg/interact"
)

const welcomeMessagePart1 = `
Welcome!

You're one of the first people to use Replicate. Isn't that exciting?

While we're still testing Replicate, we're gathering usage information and crash reports. We only store metadata on what commands you run, not any of your code or model data.

At some point we'll give you a way to opt-out of this, but if this is really not OK, email us at team@replicate.ai and we'll figure something out for you.

Please enter your email, in case we need to ask you any questions about stuff that breaks.
`

const welcomeMessagePart2 = `
Thanks!

If you have any questions or feedback, email us: team@replicate.ai

Have fun!

Andreas & Ben

~~~~~~~~~~~~~

`

// RunOnboarding gets the user's email for analytics
func RunOnboarding() (email string, err error) {
	fmt.Println(welcomeMessagePart1)

	email, err = interact.Interactive{
		Prompt:   "Email",
		Required: true,
	}.Read()

	fmt.Println(welcomeMessagePart2)

	return email, err
}
