package log

import (
	"gopkg.in/Sirupsen/logrus.v0"
)

type Fields logrus.Fields

// Like a logrus.Entry, but is nullable. This allows us to selectively disable
// logging while also removing all code overhead associated with it
type Entry struct {
	mod        Module
	lazyfields [8]func() Fields
}

func (entry Entry) log() *logrus.Entry {
	final := logrus.StandardLogger().WithField("_mod", modNames[entry.mod])
	for _, lf := range entry.lazyfields {
		if lf != nil {
			final = final.WithFields(logrus.Fields(lf()))
		}
	}

	fields := make(logrus.Fields, 8)

	var z EntryZ
	for _, c := range contexts {
		c.AddLogContext(&z)
	}
	for i := range z.zfbuf[:z.zfidx] {
		fields[z.zfbuf[i].Key] = z.zfbuf[i].Value()
	}
	return final.WithFields(fields)
}

func (entry Entry) WithFields(fields Fields) Entry {
	return entry.WithDelayedFields(func() Fields { return fields })
}

func (entry Entry) WithField(key string, value any) Entry {
	return entry.WithDelayedFields(func() Fields {
		return Fields{
			key: value,
		}
	})
}

func (entry Entry) WithDelayedFields(getfields func() Fields) Entry {
	for idx := range entry.lazyfields {
		if entry.lazyfields[idx] == nil {
			entry.lazyfields[idx] = getfields
			return entry
		}
	}
	return entry
}

func (entry Entry) Debug(args ...any) {
	if entry.mod.Enabled(DebugLevel) {
		entry.log().Debug(args...)
	}
}

func (entry Entry) Print(args ...any) {
	if entry.mod.Enabled(InfoLevel) {
		entry.log().Print(args...)
	}
}

func (entry Entry) Info(args ...any) {
	if entry.mod.Enabled(InfoLevel) {
		entry.log().Info(args...)
	}
}

func (entry Entry) Warn(args ...any) {
	if entry.mod.Enabled(WarnLevel) {
		entry.log().Warn(args...)
	}
}

func (entry Entry) Warning(args ...any) {
	if entry.mod.Enabled(WarnLevel) {
		entry.log().Warning(args...)
	}
}

func (entry Entry) Error(args ...any) {
	if entry.mod.Enabled(ErrorLevel) {
		entry.log().Error(args...)
	}
}

func (entry Entry) Fatal(args ...any) {
	if entry.mod.Enabled(FatalLevel) {
		entry.log().Fatal(args...)
	}
}

func (entry Entry) Panic(args ...any) {
	if entry.mod.Enabled(PanicLevel) {
		entry.log().Panic(args...)
	}
}

// printf-like family

func (entry Entry) Debugf(format string, args ...any) {
	if entry.mod.Enabled(DebugLevel) {
		entry.log().Debugf(format, args...)
	}
}

func (entry Entry) Printf(format string, args ...any) {
	if entry.mod.Enabled(InfoLevel) {
		entry.log().Printf(format, args...)
	}
}

func (entry Entry) Infof(format string, args ...any) {
	if entry.mod.Enabled(InfoLevel) {
		entry.log().Infof(format, args...)
	}
}

func (entry Entry) Warnf(format string, args ...any) {
	if entry.mod.Enabled(WarnLevel) {
		entry.log().Warnf(format, args...)
	}
}

func (entry Entry) Warningf(format string, args ...any) {
	if entry.mod.Enabled(WarnLevel) {
		entry.log().Warningf(format, args...)
	}
}

func (entry Entry) Errorf(format string, args ...any) {
	if entry.mod.Enabled(ErrorLevel) {
		entry.log().Errorf(format, args...)
	}
}

func (entry Entry) Fatalf(format string, args ...any) {
	if entry.mod.Enabled(FatalLevel) {
		entry.log().Fatalf(format, args...)
	}
}

func (entry Entry) Panicf(format string, args ...any) {
	if entry.mod.Enabled(PanicLevel) {
		entry.log().Panicf(format, args...)
	}
}

// New-line style family

func (entry Entry) Debugln(args ...any) {
	if entry.mod.Enabled(DebugLevel) {
		entry.log().Debugln(args...)
	}
}

func (entry Entry) Println(args ...any) {
	if entry.mod.Enabled(InfoLevel) {
		entry.log().Println(args...)
	}
}

func (entry Entry) Infoln(args ...any) {
	if entry.mod.Enabled(InfoLevel) {
		entry.log().Infoln(args...)
	}
}

func (entry Entry) Warnln(args ...any) {
	if entry.mod.Enabled(WarnLevel) {
		entry.log().Warnln(args...)
	}
}

func (entry Entry) Warningln(args ...any) {
	if entry.mod.Enabled(WarnLevel) {
		entry.log().Warningln(args...)
	}
}

func (entry Entry) Errorln(args ...any) {
	if entry.mod.Enabled(ErrorLevel) {
		entry.log().Errorln(args...)
	}
}

func (entry Entry) Fatalln(args ...any) {
	if entry.mod.Enabled(FatalLevel) {
		entry.log().Fatalln(args...)
	}
}

func (entry Entry) Panicln(args ...any) {
	if entry.mod.Enabled(PanicLevel) {
		entry.log().Panicln(args...)
	}
}
