package actor

import (
	"context"
	"log"
	"maps"
)

// Handler описывает функцию-обработчик команды:
// ctx - контекст для отмены или таймаутов,
// payload - данные сообщения.
type Handler func(ctx context.Context, payload any) error

// ActorInterface задаёт методы актора:
// Register - регистрирует обработчик для action,
// Send - отправляет сообщение с action и payload,
// Stop - останавливает цикл обработки.
type ActorInterface interface {
	Register(action string, handler Handler)
	Send(action string, payload any)
	Stop()
}

// actorImpl - приватный тип.
type actorImpl struct {
	mailbox  chan message
	handlers map[string]Handler
	cancel   context.CancelFunc
	logger   *log.Logger
	debug    bool
}

// message - внутренняя структура сообщений.
type message struct {
	action  string
	payload any
}

// NewActor создаёт и запускает актора:
// bufferSize - размер буфера очереди,
// handlersInit - начальные обработчики (можно nil),
// logger - логгер для сообщений (если nil, используется log.Default()),
// debug - включить вывод отладочных сообщений.
func NewActor(bufferSize int, handlersInit map[string]Handler, logger *log.Logger, debug bool) ActorInterface {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = log.Default() // На всякий случай.
	}
	// Приватный "класс" Actor
	a := &actorImpl{
		mailbox:  make(chan message, bufferSize),
		handlers: make(map[string]Handler),
		cancel:   cancel,
		logger:   logger,
		debug:    debug,
	}
	maps.Copy(a.handlers, handlersInit) // Копия.

	go a.start(ctx)
	return a
}

// Register создаёт или обновляет обработчик для action.
func (a *actorImpl) Register(action string, handler Handler) {
	a.handlers[action] = handler
}

// Send передает сообщение актору.
// Если action не зарегистрирован, сообщение будет проигнорировано.
func (a *actorImpl) Send(action string, payload any) {
	msg := message{action: action, payload: payload}
	select {
	case a.mailbox <- msg:
	default:
		// Блокируем при переполнении буфера.
		a.mailbox <- msg
	}
}

// Stop сигнал актору завершить работу.
func (a *actorImpl) Stop() {
	a.cancel()
}

// start обрабатывает сообщения до завершения контекста.
func (a *actorImpl) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if a.debug {
				a.logger.Println("Actor stopped")
			}
			return

		case msg := <-a.mailbox:
			handler, exists := a.handlers[msg.action]
			if !exists {
				if a.debug {
					a.logger.Printf("No handler for action %q", msg.action)
				}
				continue
			}
			if err := handler(ctx, msg.payload); err != nil {
				a.logger.Printf("Error in handler %q: %v", msg.action, err) // Можно использовать Errorf, но тут не возвращается ошибка
			}
		}
	}
}

// ------ Usage ------
//
// func main() {
//     handlers := map[string]actor.Handler{
//         "download": func(ctx context.Context, payload any) error {
//             url, ok := payload.(string)
//             if !ok {
//                 return fmt.Errorf("payload is not a string URL")
//             }
//             fmt.Println(url)
//             return nil
//         },
//     }
//
//     // Передаём log.Default() и включаем режим отладки
//     downloadActor := actor.NewActor(5, handlers, log.Default(), true)
//
//     downloadActor.Send("download", "https://example.com/file.zip")
//     downloadActor.Stop()
// }
