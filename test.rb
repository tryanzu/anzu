require 'email_reply_parser'

email_body = "Hey man, how u doing soon?\n\nFrom: pc@pedidos.spartangeek.com\nSubject: PC Spartana\nTo: fernandez14@outlook.com\nDate: Tue, 6 Oct 2015 02:57:51 +0000\n\nHola, aqui tienes tu spartana! \t\t \t   \t\t  \n"
reply = EmailReplyParser.read(email_body)

puts reply.inspect