package rp

// StageName generates a string such as
// `  => my_stage_name(["ctxVar"]) => ["newCtxVar"]`
// from the inputs StageName(true, "my_stage_name", []string{"ctxVar"}, []string{"newCtxVar"}, false)
func StageName(usesInParam bool, name string, ctxDependencies []string, ctxOutputs []string, returnsAValue bool) string {

	inArrow := "  => "
	if !usesInParam {
		inArrow = ""
	}

	params := "("
	for i, dep := range ctxDependencies {
		params += "[\"" + dep + "\"]"
		if i < len(ctxDependencies)-1 {
			params += ", "
		}
	}
	params += ")"

	funcString := name + params

	ctxOutString := ""
	if len(ctxOutputs) > 0 {
		ctxOutString = " => "
		for i, out := range ctxOutputs {
			ctxOutString += "[\"" + out + "\"]"
			if i < len(ctxOutputs)-1 {
				ctxOutString += ", "
			}
		}
	}

	outArrow := " =>"
	if !returnsAValue {
		outArrow = ""
	}

	return inArrow + funcString + ctxOutString + outArrow
}

// FuncStr generates a string such as
// `my_stage_name(["ctxVar1"], ["ctxVar2"])`
// from the inputs FuncStr("my_stage_name", "ctxVar1", "ctxVar2")
func FuncStr(name string, ctxDependencies ...string) string {

	params := "("
	for i, dep := range ctxDependencies {
		params += "[\"" + dep + "\"]"
		if i < len(ctxDependencies)-1 {
			params += ", "
		}
	}
	params += ")"

	return name + params
}

// CtxOutStr generates a string such as
// ` => ["ctxOutVar1"], ["ctxOutVar2"]`
// from the inputs CtxOutStr("ctxOutVar1", "ctxOutVar2")
func CtxOutStr(ctxOutputs ...string) string {

	if len(ctxOutputs) == 0 {
		return ""
	}

	ctxOut := " => "
	for i, out := range ctxOutputs {
		ctxOut += "[\"" + out + "\"]"
		if i < len(ctxOutputs)-1 {
			ctxOut += ", "
		}
	}

	return ctxOut
}
