package event

type Key uint8

// Based on:
// https://github.com/preactjs/preact/blob/051f10c59d14229520f14a531a4de79162e18c02/src/jsx.d.ts#L1260
const (
	// Image Events
	Load Key = iota
	Error

	// Clipboard Events
	Copy
	Cut
	Paste

	// Composition Events
	CompositionEnd
	CompositionStart
	CompositionUpdate

	// Details Events
	Toggle

	// Dialog Events
	Close
	Cancel

	// Focus Events
	Focus
	FocusIn
	FocusOut
	Blur

	// Form Events
	Change
	Input
	BeforeInput
	Search
	Submit
	Invalid
	Reset
	FormData

	// Keyboard Events
	KeyDown
	KeyPress
	KeyUp

	// Media Events
	Abort
	CanPlay
	CanPlayThrough
	DurationChange
	Emptied
	Encrypted
	Ended
	LoadedData
	LoadedMetadata
	LoadStart
	Pause
	Play
	Playing
	Progress
	RateChange
	Seeked
	Seeking
	Stalled
	Suspend
	TimeUpdate
	VolumeChange
	Waiting

	// MouseEvents
	Click
	ContextMenu
	DblClick
	Drag
	DragEnd
	DragEnter
	DragExit
	DragLeave
	DragOver
	DragStart
	Drop
	MouseDown
	MouseEnter
	MouseLeave
	MouseMove
	MouseOut
	MouseOver
	MouseUp

	// Selection Events
	Select

	// Touch Events
	TouchCancel
	TouchEnd
	TouchMove
	TouchStart

	// Pointer Events
	PointerOver
	PointerEnter
	PointerDown
	PointerMove
	PointerUp
	PointerCancel
	PointerOut
	PointerLeave

	// UI Events
	Scroll

	// Wheel Events
	Wheel

	// Animation Events
	AnimationStart
	AnimationEnd
	AnimationIteration

	// Transition Events
	TransitionCancel
	TransitionEnd
	TransitionRun
	TransitionStart

	// PictureInPicture Events
	EnterPictureInPicture
	LeavePictureInPicture
	Resize
)

// Is an event handler
func Is(key string) bool {
	_, ok := To[key]
	return ok
}

// To key from string
var To = map[string]Key{
	"onLoad":                  Load,
	"onError":                 Error,
	"onCopy":                  Copy,
	"onCut":                   Cut,
	"onPaste":                 Paste,
	"onCompositionEnd":        CompositionEnd,
	"onCompositionStart":      CompositionStart,
	"onCompositionUpdate":     CompositionUpdate,
	"onToggle":                Toggle,
	"onClose":                 Close,
	"onCancel":                Cancel,
	"onFocus":                 Focus,
	"onFocusIn":               FocusIn,
	"onFocusOut":              FocusOut,
	"onBlur":                  Blur,
	"onChange":                Change,
	"onInput":                 Input,
	"onBeforeInput":           BeforeInput,
	"onSearch":                Search,
	"onSubmit":                Submit,
	"onInvalid":               Invalid,
	"onReset":                 Reset,
	"onFormData":              FormData,
	"onKeyDown":               KeyDown,
	"onKeyPress":              KeyPress,
	"onKeyUp":                 KeyUp,
	"onAbort":                 Abort,
	"onCanPlay":               CanPlay,
	"onCanPlayThrough":        CanPlayThrough,
	"onDurationChange":        DurationChange,
	"onEmptied":               Emptied,
	"onEncrypted":             Encrypted,
	"onEnded":                 Ended,
	"onLoadedData":            LoadedData,
	"onLoadedMetadata":        LoadedMetadata,
	"onLoadStart":             LoadStart,
	"onPause":                 Pause,
	"onPlay":                  Play,
	"onPlaying":               Playing,
	"onProgress":              Progress,
	"onRateChange":            RateChange,
	"onSeeked":                Seeked,
	"onSeeking":               Seeking,
	"onStalled":               Stalled,
	"onSuspend":               Suspend,
	"onTimeUpdate":            TimeUpdate,
	"onVolumeChange":          VolumeChange,
	"onWaiting":               Waiting,
	"onClick":                 Click,
	"onContextMenu":           ContextMenu,
	"onDblClick":              DblClick,
	"onDrag":                  Drag,
	"onDragEnd":               DragEnd,
	"onDragEnter":             DragEnter,
	"onDragExit":              DragExit,
	"onDragLeave":             DragLeave,
	"onDragOver":              DragOver,
	"onDragStart":             DragStart,
	"onDrop":                  Drop,
	"onMouseDown":             MouseDown,
	"onMouseEnter":            MouseEnter,
	"onMouseLeave":            MouseLeave,
	"onMouseMove":             MouseMove,
	"onMouseOut":              MouseOut,
	"onMouseOver":             MouseOver,
	"onMouseUp":               MouseUp,
	"onSelect":                Select,
	"onTouchCancel":           TouchCancel,
	"onTouchEnd":              TouchEnd,
	"onTouchMove":             TouchMove,
	"onTouchStart":            TouchStart,
	"onPointerOver":           PointerOver,
	"onPointerEnter":          PointerEnter,
	"onPointerDown":           PointerDown,
	"onPointerMove":           PointerMove,
	"onPointerUp":             PointerUp,
	"onPointerCancel":         PointerCancel,
	"onPointerOut":            PointerOut,
	"onPointerLeave":          PointerLeave,
	"onScroll":                Scroll,
	"onWheel":                 Wheel,
	"onAnimationStart":        AnimationStart,
	"onAnimationEnd":          AnimationEnd,
	"onAnimationIteration":    AnimationIteration,
	"onTransitionCancel":      TransitionCancel,
	"onTransitionEnd":         TransitionEnd,
	"onTransitionRun":         TransitionRun,
	"onTransitionStart":       TransitionStart,
	"onEnterPictureInPicture": EnterPictureInPicture,
	"onLeavePictureInPicture": LeavePictureInPicture,
	"onResize":                Resize,
}

// From key to string
var From = map[Key]string{
	Load:                  "onLoad",
	Error:                 "onError",
	Copy:                  "onCopy",
	Cut:                   "onCut",
	Paste:                 "onPaste",
	CompositionEnd:        "onCompositionEnd",
	CompositionStart:      "onCompositionStart",
	CompositionUpdate:     "onCompositionUpdate",
	Toggle:                "onToggle",
	Close:                 "onClose",
	Cancel:                "onCancel",
	Focus:                 "onFocus",
	FocusIn:               "onFocusIn",
	FocusOut:              "onFocusOut",
	Blur:                  "onBlur",
	Change:                "onChange",
	Input:                 "onInput",
	BeforeInput:           "onBeforeInput",
	Search:                "onSearch",
	Submit:                "onSubmit",
	Invalid:               "onInvalid",
	Reset:                 "onReset",
	FormData:              "onFormData",
	KeyDown:               "onKeyDown",
	KeyPress:              "onKeyPress",
	KeyUp:                 "onKeyUp",
	Abort:                 "onAbort",
	CanPlay:               "onCanPlay",
	CanPlayThrough:        "onCanPlayThrough",
	DurationChange:        "onDurationChange",
	Emptied:               "onEmptied",
	Encrypted:             "onEncrypted",
	Ended:                 "onEnded",
	LoadedData:            "onLoadedData",
	LoadedMetadata:        "onLoadedMetadata",
	LoadStart:             "onLoadStart",
	Pause:                 "onPause",
	Play:                  "onPlay",
	Playing:               "onPlaying",
	Progress:              "onProgress",
	RateChange:            "onRateChange",
	Seeked:                "onSeeked",
	Seeking:               "onSeeking",
	Stalled:               "onStalled",
	Suspend:               "onSuspend",
	TimeUpdate:            "onTimeUpdate",
	VolumeChange:          "onVolumeChange",
	Waiting:               "onWaiting",
	Click:                 "onClick",
	ContextMenu:           "onContextMenu",
	DblClick:              "onDblClick",
	Drag:                  "onDrag",
	DragEnd:               "onDragEnd",
	DragEnter:             "onDragEnter",
	DragExit:              "onDragExit",
	DragLeave:             "onDragLeave",
	DragOver:              "onDragOver",
	DragStart:             "onDragStart",
	Drop:                  "onDrop",
	MouseDown:             "onMouseDown",
	MouseEnter:            "onMouseEnter",
	MouseLeave:            "onMouseLeave",
	MouseMove:             "onMouseMove",
	MouseOut:              "onMouseOut",
	MouseOver:             "onMouseOver",
	MouseUp:               "onMouseUp",
	Select:                "onSelect",
	TouchCancel:           "onTouchCancel",
	TouchEnd:              "onTouchEnd",
	TouchMove:             "onTouchMove",
	TouchStart:            "onTouchStart",
	PointerOver:           "onPointerOver",
	PointerEnter:          "onPointerEnter",
	PointerDown:           "onPointerDown",
	PointerMove:           "onPointerMove",
	PointerUp:             "onPointerUp",
	PointerCancel:         "onPointerCancel",
	PointerOut:            "onPointerOut",
	PointerLeave:          "onPointerLeave",
	Scroll:                "onScroll",
	Wheel:                 "onWheel",
	AnimationStart:        "onAnimationStart",
	AnimationEnd:          "onAnimationEnd",
	AnimationIteration:    "onAnimationIteration",
	TransitionCancel:      "onTransitionCancel",
	TransitionEnd:         "onTransitionEnd",
	TransitionRun:         "onTransitionRun",
	TransitionStart:       "onTransitionStart",
	EnterPictureInPicture: "onEnterPictureInPicture",
	LeavePictureInPicture: "onLeavePictureInPicture",
	Resize:                "onResize",
}
