package server

// Response is the response returned by a module call
type Response struct {
	Success bool
	// This should be set to true if the work caused any changes
	Modified bool
	FuncName string
	ErrorMsg string
	Data     []byte
}

// Append adds the given work response to this one
func (w *Response) Append(rsp Response) {
	if !rsp.Success {
		w.Success = false
	}
	if rsp.Modified {
		w.Modified = true
	}
	if len(rsp.ErrorMsg) > 0 {
		if len(w.ErrorMsg) > 0 {
			w.ErrorMsg = "\n" + rsp.ErrorMsg
		} else {
			w.ErrorMsg = rsp.ErrorMsg
		}
	}
	w.Data = append(w.Data, rsp.Data...)
}

// ModuleMgr deals with calling modules
type ModuleMgr interface {
	// SendRequest to a module function
	SendRequest(session, module, funcName string, data []byte) (Response, error)
	Close()
}
