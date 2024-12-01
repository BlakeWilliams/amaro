// Package flash extends the amaro session middleware to provide flash
// messages. Flash messages are messages stored in the session that are deleted
// after being accessed once.
//
// To use, simply add `*flash.Messages` to your session data, and use the Middleware in your app.
//
//	type SessionData struct {
//	    Flash *flash.Messages
//	}
package flash
