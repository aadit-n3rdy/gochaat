package common

type Message struct {
	MsgType string;
	From string;
	To string;
	Data []byte;
};

const MSG_HELLO = "CC_MSG_HELLO";
const MSG_TEXT = "CC_MSG_TEXT";
const MSG_CERT_REQ = "CC_MSG_CERT_REQ";
const MSG_CERT_RES = "CC_MSG_CERT_RES";
