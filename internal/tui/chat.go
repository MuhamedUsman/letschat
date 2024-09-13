package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"log/slog"
)

const (
	chatMenu     = "chatMenu"
	chatViewport = "chatViewport"
	chatTxtarea  = "chatTxtarea"
)

type ChatModel struct {
	chatTxtarea  textarea.Model
	chatViewport viewport.Model
	n            int
	focusIdx     int // 0 -> chatTxtarea, 1 -> chatViewport
	focus        bool
}

func InitialChatModel() ChatModel {
	return ChatModel{
		chatTxtarea:  newChatTxtArea(),
		chatViewport: newChatViewport(0, 0),
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	if !m.focus {
		m.chatTxtarea.Blur()
		m.updateChatTxtareaAndViewportDimensions()
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateChatTxtareaAndViewportDimensions()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+t":
			cmd := m.chatTxtarea.Focus()
			m.updateChatTxtareaAndViewportDimensions()
			return m, cmd
		case "ctrl+s":
			m.chatTxtarea.Reset()
			m.chatTxtarea.Blur()
			m.chatTxtarea.SetHeight(2) // we initialize with 0
		case "esc":
			m.chatTxtarea.Blur()
			return m, func() tea.Msg { return xMsg{} }
		}
		//m.updateChatTxtareaAndViewportDimensions()
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if zone.Get("helllo").InBounds(msg) {
				return m, tea.Quit
			}
		case tea.MouseButtonWheelDown:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.CursorDown()
			}
		case tea.MouseButtonWheelUp:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.CursorUp()
			}
		case tea.MouseButtonWheelRight:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.SetCursor(m.chatTxtarea.LineInfo().CharOffset + 1)
			}
		case tea.MouseButtonWheelLeft:
			if zone.Get(chatTxtarea).InBounds(msg) {
				m.chatTxtarea.SetCursor(max(0, m.chatTxtarea.LineInfo().CharOffset-1))
			}
		default:
		}
		if zone.Get(chatViewport).InBounds(msg) {
			m.chatTxtarea.Blur()
			m.updateChatTxtareaAndViewportDimensions()
		}
		if zone.Get(chatTxtarea).InBounds(msg) {
			cmd := m.chatTxtarea.Focus()
			m.updateChatTxtareaAndViewportDimensions()
			return m, cmd
		}
	case xMsg:
		m.n = 0
		slog.Info("What is your name")
	}
	m.chatViewport.SetContent(renderChatViewport())
	return m, tea.Batch(m.handleChatTextareaUpdate(msg), m.handleChatViewportUpdate(msg))
}

func (m ChatModel) View() string {
	h := renderChatHeader("Muhammad Usman")
	chatHeaderHeight = lipgloss.Height(h)
	ta := zone.Mark(chatTxtarea, m.chatTxtarea.View())
	ta = renderChatTextarea(ta, m.chatTxtarea.Focused())
	chatTextareaHeight = lipgloss.Height(ta)
	m.chatViewport.Height = chatHeight() - (chatHeaderHeight + chatTextareaHeight)
	//chatViewportHeight := chatHeight() - (lipgloss.Height(h) + lipgloss.Height(ta))
	/*m1 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(chatBubbleLStyle.Width(chatWidth() - 20).Render("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."))
	m2 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Right).Render(chatBubbleRStyle.Width(chatWidth() - 20).Render("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc."))
	m3 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(chatBubbleLStyle.Width(chatWidth() - 20).Render("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."))
	content := lipgloss.JoinVertical(lipgloss.Center, m1, m2, m3)*/
	//chatView := renderChatViewport(content, chatViewportHeight)
	//m.chatViewport.SetContent(content)
	chatView := m.chatViewport.View()
	chatView = zone.Mark(chatViewport, chatView)
	c := lipgloss.JoinVertical(lipgloss.Top, h, chatView, ta)
	c = chatContainerStyle.
		Width(chatWidth()).
		Height(chatHeight()).
		Render(c)
	return c
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func newChatTxtArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Prompt = ""
	ta.CharLimit = 1000
	ta.ShowLineNumbers = false
	ta.SetHeight(0)
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Foreground(whiteColor)
	return ta
}

func newChatViewport(w, h int) viewport.Model {
	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true
	return vp
}

func renderChatHeader(name string) string {
	c := chatHeaderStyle.Width(chatWidth())
	menu := zone.Mark(chatMenu, "⚙️")
	menuMarginLeft := max(0, c.GetWidth()-(c.GetHorizontalFrameSize()+lipgloss.Width(name)+lipgloss.Width(menu)))
	menu = lipgloss.NewStyle().
		MarginLeft(menuMarginLeft).
		Render(menu)
	return c.Render(name, menu)
}

func renderChatViewport() string {
	m1 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(chatBubbleLStyle.Width(chatWidth() - 20).Render("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi a rutrum purus. Morbi lacinia suscipit elit ac luctus. Ut quis tempus nisi. Ut pulvinar purus vitae mauris venenatis malesuada. Mauris scelerisque odio purus, ac ornare quam convallis a. Cras lacinia libero arcu, vel interdum magna tincidunt id. Integer quis pulvinar mi. Donec accumsan molestie odio quis tempor. Curabitur id pellentesque ligula. Maecenas placerat ex non lorem consectetur, ac suscipit nisi hendrerit. Sed id enim ex.\n\nVestibulum mattis, tortor vel scelerisque feugiat, lorem mi ullamcorper justo, vel dictum orci felis vitae libero. Ut sodales hendrerit consectetur. Cras felis augue, vehicula vel justo nec, fringilla ultricies metus. Pellentesque non gravida leo. Suspendisse at arcu ac odio volutpat dignissim. Cras eu massa at eros vulputate euismod quis quis dui. Vivamus imperdiet non orci pharetra ultricies. Nulla condimentum nibh eu nisl bibendum, sed pharetra est cursus. Proin rutrum fringilla efficitur.\n\nNulla facilisi. Fusce massa lectus, accumsan et blandit in, elementum rhoncus nulla. Morbi cursus, eros id placerat fringilla, sapien risus pharetra leo, nec scelerisque sapien dui a nibh. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nullam ac ipsum sem. Vestibulum congue fermentum aliquet. Quisque non elit lectus. Phasellus euismod varius mauris, et ultricies nulla dignissim nec. Curabitur euismod posuere metus sed feugiat. Sed sed commodo erat.\n\nOrci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Phasellus lacus lacus, rhoncus at imperdiet eget, sodales a risus. Cras nec cursus tellus. Maecenas sagittis sollicitudin volutpat. Nulla quis elementum sapien. Aliquam et nisi in libero sodales accumsan. Sed hendrerit sem at consectetur dignissim. Etiam varius ultricies eros vel faucibus. Maecenas vehicula felis nec mi ullamcorper posuere. Aliquam porttitor tempus mattis. Vestibulum quis eleifend dui. Maecenas tincidunt enim in ultrices elementum. Duis nec odio vel felis venenatis imperdiet vitae et mi. Sed in quam convallis, bibendum sapien et, molestie felis. Nulla vitae velit dolor.\n\nProin eu venenatis arcu. Maecenas condimentum nisl sollicitudin ante facilisis convallis. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Suspendisse leo urna, auctor at justo id, volutpat imperdiet mi. Vestibulum ut hendrerit orci, vel vulputate orci. Suspendisse pulvinar velit placerat nibh tempus venenatis a at sapien. Nam eget ante tellus. Sed suscipit neque quis dapibus tincidunt.\n\nIn dictum aliquam maximus. Nunc maximus eu leo vitae vulputate. In nisi tellus, lobortis vitae turpis eget, molestie lacinia eros. Mauris malesuada, felis sit amet molestie volutpat, mauris mauris ultrices felis, sit amet rhoncus sem enim at urna. Sed nisi risus, ultricies eu orci ac, tempus fringilla quam. Aliquam vitae porta sem. Suspendisse potenti. Curabitur lacus justo, posuere sit amet imperdiet sit amet, blandit id lorem. Suspendisse et nisl id erat suscipit tempus ac at ex. Aliquam vel nisi imperdiet, venenatis ex in, condimentum magna. Morbi elementum molestie ligula, vel rutrum elit aliquam non. Suspendisse at lorem mauris.\n\nDonec sem odio, euismod vestibulum sem nec, porttitor fermentum diam. Integer in nisi auctor, convallis dolor at, vulputate tellus. Donec sit amet varius lectus. Duis a lobortis augue. Donec eget molestie leo, non porttitor quam. Duis vel odio ipsum. Cras vel orci bibendum, porttitor felis non, tempor neque. Praesent non purus a tortor condimentum lacinia. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Nam tortor enim, venenatis eu rutrum vel, suscipit eget nunc. Sed commodo ac dolor id consequat. Quisque et purus tincidunt, scelerisque ligula nec, varius turpis. Maecenas ut finibus odio.\n\nVivamus at sem quis eros commodo porta non a mi. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse potenti. In hac habitasse platea dictumst. Aliquam at velit sed ligula venenatis tempus. Praesent metus urna, porttitor sit amet nibh vel, gravida iaculis neque. Nam a dapibus turpis. Nullam ultrices ante sed bibendum facilisis. Sed eget porttitor nisl, et eleifend magna. Ut nec tincidunt neque. Donec posuere eleifend nunc, sit amet convallis diam consectetur eget. Vivamus tempor tincidunt dolor ac lacinia. Donec nec ex ac diam fermentum vehicula. Integer non lacus diam.\n\nOrci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Morbi faucibus aliquam dui laoreet convallis. Proin ut tincidunt nisl. In nisl est, feugiat non mollis a, egestas ut nisl. Morbi accumsan quam nec urna ornare, eget fringilla ligula imperdiet. Vivamus faucibus, nibh sed consectetur malesuada, neque libero semper metus, ac volutpat leo tortor nec lacus. Maecenas ut ultricies odio. Ut in faucibus nisi, at pharetra mauris. Mauris est nisi, ultricies eget turpis vitae, posuere commodo ligula. Cras tempus luctus tortor, eget blandit risus dictum vitae. Phasellus erat purus, interdum et quam ac, suscipit malesuada dolor. Sed iaculis pulvinar enim, vel luctus odio mollis ut. Praesent sem velit, congue sit amet sem at, accumsan porttitor mi. Nam porttitor felis non congue ornare.\n\nMaecenas posuere vel est non eleifend. Aliquam sagittis scelerisque felis. Phasellus aliquam lorem massa, et cursus urna sollicitudin at. In pharetra ipsum leo, quis malesuada dolor efficitur a. Nam mattis efficitur felis, quis euismod leo fringilla ac. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Praesent ut nibh ac massa commodo tempus vel sed mauris. Sed feugiat vestibulum risus sed mattis. Mauris lacus nunc, facilisis a magna quis, blandit pellentesque ipsum. Cras mollis aliquam tincidunt. Vivamus tristique placerat diam et luctus. Donec nec odio euismod orci dictum hendrerit. Sed lobortis tincidunt ipsum, a vehicula erat mollis in. Nam non aliquet nisl. Quisque erat elit, elementum in dui sit amet, facilisis bibendum augue.\n\nProin eget tortor id ipsum pretium porttitor. Sed velit dolor, porttitor id elit vel, ullamcorper condimentum enim. Phasellus finibus lacus ut massa volutpat vehicula. Suspendisse dapibus nisl erat, at hendrerit metus fermentum a. Curabitur malesuada magna eget tortor porttitor pellentesque. In at viverra lorem. Nulla erat magna, efficitur vitae pharetra nec, gravida a mauris. Etiam sit amet hendrerit leo, id ultricies libero. Etiam massa erat, placerat ut sem ut, tristique pharetra dui. Sed augue urna, feugiat non condimentum ac, vehicula sit amet enim. Duis eget semper dolor. Morbi dignissim massa aliquam, imperdiet tellus vel, sagittis nibh. Nam eget."))
	m2 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Right).Render(chatBubbleRStyle.Width(chatWidth() - 20).Render("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc."))
	m3 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(chatBubbleLStyle.Width(chatWidth() - 20).Render("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."))
	m4 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(chatBubbleLStyle.Width(chatWidth() - 20).Render("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."))
	m5 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Right).Render(chatBubbleRStyle.Width(chatWidth() - 20).Render("There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc."))
	m6 := chatBubbleContainer.Width(chatWidth() - chatBubbleContainer.GetHorizontalFrameSize()).Align(lipgloss.Left).Render(chatBubbleLStyle.Width(chatWidth() - 20).Render("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."))

	m3 = zone.Mark("helllo", m3)
	content := lipgloss.JoinVertical(lipgloss.Center, m1, m2, m3, m4, m5, m6)
	return content
}

func renderChatTextarea(ta string, padding bool) string {
	cStyle := chatTxtareaStyle.
		Width(chatWidth())
	if padding {
		cStyle = cStyle.UnsetPadding()
		cStyle = cStyle.Height(5)
	}
	return cStyle.Render(ta)
}

func (m *ChatModel) handleChatTextareaUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.chatTxtarea, cmd = m.chatTxtarea.Update(msg)
	return cmd
}

func (m *ChatModel) handleChatViewportUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.chatViewport, cmd = m.chatViewport.Update(msg)
	return cmd
}

func (m *ChatModel) updateChatTxtareaAndViewportDimensions() {
	m.chatTxtarea.SetWidth(chatWidth() - chatTxtareaStyle.GetHorizontalFrameSize())
	if m.chatTxtarea.Focused() {
		m.chatTxtarea.SetHeight(5)
	} else {
		m.chatTxtarea.SetHeight(0)
	}
	m.chatViewport.Width = chatWidth()
	m.chatViewport.Height = chatHeight() - (chatHeaderHeight + chatTextareaHeight)
}

type xMsg struct{}
