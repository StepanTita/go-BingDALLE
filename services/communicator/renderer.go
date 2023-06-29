package communicator

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"

	"github.com/StepanTita/go-BingDALLE/config"
	"github.com/StepanTita/go-BingDALLE/dalle"
	"github.com/StepanTita/go-BingDALLE/terminal"
)

type renderer struct {
	log *logrus.Entry
	ctx context.Context
	cfg config.Config

	userInput string
	content   string

	bot dalle.Bot

	parsedResponsesChan <-chan dalle.PollStatus

	// tea Model fields
	prefix string

	options []tea.ProgramOption

	program  *tea.Program
	Error    *communicatorError
	state    state
	styles   terminal.Styles
	renderer *lipgloss.Renderer
	anim     tea.Model

	width  int
	height int
}

// chatInput is a tea.Msg that wraps the content read from stdin.
type chatInput struct {
	content string
}

// chatOutput a tea.Msg that wraps the content returned from openai.
type chatOutput struct {
	content string
}

func newRenderer(cfg config.Config, r *lipgloss.Renderer, opts ...tea.ProgramOption) *renderer {
	styles := terminal.NewStyles(r)

	prefix := styles.Prefix.Render("DaLLE >")
	if r.ColorProfile() == termenv.TrueColor {
		prefix = terminal.MakeGradientText(styles.Prefix, "DaLLE >")
	}

	rend := &renderer{
		log: cfg.Logging().WithField("service", "[RENDERER]"),
		cfg: cfg,

		prefix: prefix,

		state:    startState,
		renderer: r,
		styles:   styles,

		bot: dalle.New(cfg),

		options: opts,
	}

	rend.program = tea.NewProgram(rend, opts...)
	return rend
}

func (r *renderer) withContext(ctx context.Context) *renderer {
	r.ctx = ctx
	return r
}

func (r *renderer) withState(state state) *renderer {
	r.state = state
	return r
}

func (r *renderer) withInput(input string) *renderer {
	r.userInput = input
	return r
}

func (r *renderer) withContent(content string) *renderer {
	r.content = content
	return r
}

func (r *renderer) run(ctx context.Context) error {
	*r = *r.withContext(ctx)

	r.program = tea.NewProgram(r, r.options...)

	_, err := r.program.Run()

	return err
}

// Init implements tea.Model.
func (r *renderer) Init() tea.Cmd {
	var err error
	r.parsedResponsesChan, err = r.bot.CreateImages(r.ctx, r.userInput)
	if err != nil {
		r.log.WithError(err).Error("failed to ask bot")
		return func() tea.Msg {
			return communicatorError{
				err:    err,
				reason: "failed to ask bot",
			}
		}
	}
	return func() tea.Msg {
		return chatInput{}
	}
}

// Update implements tea.Model.
func (r *renderer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case chatInput:
		if r.state == startState {
			r.anim = terminal.NewCyclingChars(cyclingChars, generationText, r.renderer, r.styles)
			r.state = completionState
		}
		r.content = msg.content
		return r, tea.Batch(r.anim.Init(), r.readResponse)
	case chatOutput:
		return r, tea.Quit
	case communicatorError:
		r.Error = &msg
		r.state = errorState
		return r, tea.Quit
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+с":
			return r, tea.Quit
		}
	}
	if r.state == completionState {
		var cmd tea.Cmd
		r.anim, cmd = r.anim.Update(msg)
		return r, cmd
	}
	return r, nil
}

// View implements tea.Model.
func (r *renderer) View() string {
	switch r.state {
	case errorState:
		return r.ErrorView()
	case completionState:
		return r.anim.View()
	}
	return r.FormattedOutput()
}

// ErrorView renders the currently set modsError
func (r *renderer) ErrorView() string {
	const maxWidth = 120
	const horizontalPadding = 2
	w := r.width - (horizontalPadding * 2)
	if w > maxWidth {
		w = maxWidth
	}
	s := r.renderer.NewStyle().Width(w).Padding(0, horizontalPadding)
	return fmt.Sprintf(
		"\n%s\n\n%s\n\n",
		s.Render(r.styles.ErrorHeader.String(), r.Error.reason),
		s.Render(r.styles.ErrorDetails.Render(r.Error.Error())),
	)
}

// FormattedOutput returns the response from OpenAI with the user configured
// prefix and standard in settings.
func (r *renderer) FormattedOutput() string {
	return r.content + "\n"
}

// readResponse reads single frame from the input channel
func (r *renderer) readResponse() tea.Msg {
	parsedFrame, ok := <-r.parsedResponsesChan
	if parsedFrame.Err != nil {
		return communicatorError{
			err:    parsedFrame.Err,
			reason: "failed to read poll",
		}
	}

	if !ok {
		r.state = completionDoneState
		return chatOutput{}
	}
	return chatInput{content: strings.Join(parsedFrame.Links, "\n")}
}
