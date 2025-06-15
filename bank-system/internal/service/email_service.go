package service

import (
	"fmt"
	"time"

	"gopkg.in/gomail.v2"

	"bank-service/internal/config"
)

type EmailService interface {
	SendCreditApprovalEmail(userID int64, amount float64, interestRate float64, monthlyPayment float64, term int) error
	SendPaymentSuccessEmail(userID int64, amount float64, creditID int64) error
	SendPaymentOverdueEmail(userID int64, amount float64, creditID int64) error
}

type emailService struct {
	config config.SMTPConfig
}

func NewEmailService(config config.SMTPConfig) EmailService {
	return &emailService{
		config: config,
	}
}

func (s *emailService) SendCreditApprovalEmail(userID int64, amount float64, interestRate float64, monthlyPayment float64, term int) error {
	subject := "Ваш кредит одобрен!"
	body := fmt.Sprintf(`
		<h1>Поздравляем! Ваш кредит одобрен</h1>
		<p>Детали кредита:</p>
		<ul>
			<li>Сумма: %.2f руб.</li>
			<li>Процентная ставка: %.2f%%</li>
			<li>Ежемесячный платеж: %.2f руб.</li>
			<li>Срок: %d месяцев</li>
		</ul>
		<p>Средства уже зачислены на ваш счет.</p>
		<p>С уважением, Ваш Банк</p>
	`, amount, interestRate, monthlyPayment, term)

	userEmail := "user@example.com"

	return s.sendEmail(userEmail, subject, body)
}

func (s *emailService) SendPaymentSuccessEmail(userID int64, amount float64, creditID int64) error {
	subject := "Платеж по кредиту выполнен успешно"
	body := fmt.Sprintf(`
		<h1>Платеж по кредиту выполнен успешно</h1>
		<p>Детали платежа:</p>
		<ul>
			<li>Сумма платежа: %.2f руб.</li>
			<li>Номер кредита: %d</li>
			<li>Дата платежа: %s</li>
		</ul>
		<p>Спасибо за своевременную оплату!</p>
		<p>С уважением, Ваш Банк</p>
	`, amount, creditID, time.Now().Format("02.01.2006"))

	userEmail := "user@example.com"

	return s.sendEmail(userEmail, subject, body)
}

func (s *emailService) SendPaymentOverdueEmail(userID int64, amount float64, creditID int64) error {
	subject := "Важно: Просрочка платежа по кредиту"
	body := fmt.Sprintf(`
		<h1>Уведомление о просрочке платежа</h1>
		<p>Уважаемый клиент,</p>
		<p>Сообщаем вам о просрочке платежа по кредиту №%d.</p>
		<p>Детали:</p>
		<ul>
			<li>Сумма платежа: %.2f руб.</li>
			<li>Дата платежа: %s</li>
		</ul>
		<p>На сумму просроченного платежа будет начислен штраф в размере 10%%.</p>
		<p>Пожалуйста, пополните счет для погашения задолженности.</p>
		<p>С уважением, Ваш Банк</p>
	`, creditID, amount, time.Now().Format("02.01.2006"))

	userEmail := "user@example.com"

	return s.sendEmail(userEmail, subject, body)
}

func (s *emailService) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.Host, 587, s.config.Username, s.config.Password)

	return d.DialAndSend(m)
}
