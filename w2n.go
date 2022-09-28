package fsring

type Write2Next func(data []byte) (wrote int, e error)

func Write2NextBuilderNew(w Write) func(Next) func(dirname string) Write2Next {
	return func(n Next) func(string) Write2Next {
		return func(dirname string) Write2Next {
			f, _ := ComposeErr(
				n,                       // dirname string -> next string, e error
				ErrFuncGen(CurryErr(w)), // string -> func([]byte)(int, error), error
			)(dirname)
			return f
		}
	}
}
