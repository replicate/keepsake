package analytics

import (
	"fmt"

	"replicate.ai/cli/pkg/interact"
)

const welcomeMessagePart1 = `
Welcome!

You're one of the first people to use Replicate. Isn't that exciting?

First -- thank you. We hope Replicate is going to make your life easier. But, it's still kinda janky. Your patience and feedback is what will make this into a really good tool.

Second -- a favor. While we're still testing it out, we want to gather usage information and crash reports so we can see what's working and what isn't.

We're only storing metadata about what commands you run, not any of your code or model data. At some point we'll make it possible to opt-out of this, but at this early stage we're gathering data from everybody. If this is really not OK, email us at team@replicate.ai and we'll figure something out for you.

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
