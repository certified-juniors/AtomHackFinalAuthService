package smtp

import (
	"fmt"
	"github.com/certified-juniors/AtomHack/internal/domain"
	"github.com/joho/godotenv"
	"net/smtp"
	"os"
	"strconv"
)

type Sender struct {
}

func NewSMTP() *Sender {
	return &Sender{}
}

func customAuth(username, password, host string) smtp.Auth {
	return &loginAuth{username, password, host}
}

type loginAuth struct {
	username, password, host string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unexpected server challenge: %s", string(fromServer))
		}
	}
	return nil, nil
}

func getParams() domain.SMTPParams {
	params := domain.SMTPParams{}

	port, _ := strconv.Atoi(os.Getenv("EMAIL_SERVICE_SMTP_PORT"))
	params.SmtpHost = os.Getenv("MAIL_SERVICE_SMTP_HOST")
	params.SmtpPort = port
	paraNoreplyUsername := os.Getenv("EMAIL_SERVICE_SMTP_NO_REPLY_USERNAME")
	NoreplyPassword := os.Getenv("EMAIL_SERVICE_SMTP_NO_REPLY_PASSWORD")
	SupportUsername := os.Getenv("EMAIL_SERVICE_SMTP_SUPPORT_USERNAME")
	SupportPassword := os.Getenv("EMAIL_SERVICE_SMTP_SUPPORT_PASSWORD")
}

func (s *Sender) SendMailToSupport(subject, body string) error {
	params := getParams()
	conn, err := smtp.Dial(fmt.Sprintf("%s:%d", params.SmtpHost, params.SmtpPort))
	if err != nil {
		fmt.Println("Ошибка при подключении к SMTP серверу:", err)
		return err
	}
	defer conn.Close()

	// Аутентификация с использованием кастомной функции
	auth := customAuth(params.NoreplyUsername, params.NoreplyPassword, params.SmtpHost)
	if err := conn.Auth(auth); err != nil {
		fmt.Println("Ошибка при аутентификации:", err)
		return err
	}

	// Отправка письма
	if err := conn.Mail(params.NoreplyUsername); err != nil {
		fmt.Println("Ошибка при отправке адреса отправителя:", err)
		return err
	}
	//if err := conn.Rcpt(s.SMTPCfg.SupportUsername); err != nil {
	//	fmt.Println("Ошибка при отправке адреса получателя:", err)
	//	return err
	//}
	data, err := conn.Data()
	if err != nil {
		fmt.Println("Ошибка при отправке данных письма:", err)
		return err
	}
	defer data.Close()

	// Формирование заголовков письма
	message := fmt.Sprintf("From: %s\r\n", params.NoreplyUsername)
	message += fmt.Sprintf("To: %s\r\n", params.SupportUsername)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-version: 1.0\r\n"
	message += "Content-Type: multipart/mixed; boundary=boundary\r\n\r\n"
	message += "--boundary\r\n"
	message += "Content-Type: text/plain; charset=utf-8\r\n"
	message += "\r\n" + body + "\r\n"

	// Добавление вложений к письму
	//for filename, content := range attachments {
	//	message += fmt.Sprintf("--boundary\r\n")
	//	message += fmt.Sprintf("Content-Type: application/octet-stream\r\n")
	//	message += fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename)
	//	message += "Content-Transfer-Encoding: base64\r\n\r\n"
	//	// Кодирование содержимого файла в Base64
	//	encoded := base64.StdEncoding.EncodeToString(content)
	//	message += encoded + "\r\n"
	//}

	// Завершение письма
	message += "--boundary--\r\n"

	// Отправка письма
	_, err = fmt.Fprintf(data, message)
	if err != nil {
		fmt.Println("Ошибка при записи данных письма:", err)
		return err
	}

	fmt.Println("Письмо успешно отправлено!")
	return nil
}

func main() {
	_ = godotenv.Load()

	s := NewSMTP()
	s.SendMailToSupport("localhost", "somebody")
}
