# vkbot
## Установка и обновление
`go get -u github.com/azzzak/vkbot`
## Запуск
Для работы бота потребуется указать адрес для отправки callback-запросов на вкладке `Управление сообщестом > Работа с API > Callback API`. Для подтверждения нового сервера надо скопировать строку подтверждения и передать ее боту. После запуска бота нажмите кнопку `Подтвердить`, в случае успеха можно будет выбрать какие типы событий станут инициировать callback-запрос.

Поддерживаются следующие типы событий:
* Входящие сообщения
* Исходящие сообщения
* Вступление в сообщество
* Выход из сообщества 

Для отправки сообщений необходимо создать ключи доступа на вкладке `Управление сообщестом > Работа с API > Ключи доступа`. Ключ указывается при запуске бота вместе с числовым идентификатором группы.

На вкладке `Управление сообщестом > Работа с API > Callback API` можно задать секретный ключ, который будет удостоверять подлинность данных. Если секретный ключ задан, его так же требуется указать при запуске бота.
## Работа за Nginx
Для работы за Nginx надо настроить обратный прокси. Функцию `ListenForWebhook` в этом случае следует запускать с параметром `"/"`.

```
server {
...
	location = /callback/path {
		proxy_pass	http://localhost:8101;
	}
...	
}

```
## Пример
Простой бот, который повторяет присланные сообщения.
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/azzzak/vkbot"
)

func main() {
	bot, err := vkbot.NewBotAPI("key", GroupID)
	if err != nil {
		fmt.Println(err)
	}
	bot.Confirmation = "123"
	bot.Secret = "123"
	updates := bot.ListenForWebhook("/callback/path")

	go http.ListenAndServe(":8101", nil)

	for update := range updates {
		switch update.Type {
		case vkbot.IncomingMessage:
      			bot.Send(update.Payload.UserID, update.Payload.Body)
		}
	}
}

```
