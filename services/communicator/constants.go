package communicator

const (
	// TODO: implement help command
	HelpCommand  = "!help"
	ExitCommand  = "!exit"
	ResetCommand = "!reset"
)

const cyclingChars = 9
const (
	generationText = "Generation"
)

type state int

const (
	startState state = iota
	completionState
	completionDoneState
	errorState
)
