package server

import "github.com/shuhuNB515/Argus/config"

// 面向用户的后端消息(中文字符原文)替换为活动输出语言的目录。. 在这个存储库中
// writeErr 回应短语·一些 writeJSON error字段是中文字符常量·它被写成字面意思(F3b/F34).
// 这里，使用中文字符原文作为密钥。, 英语·中国人·将英语翻译设置为值. 中文(ko)是
// 因为它本身就是密钥，不存储它, 查找失败·ko 在某些情况下，原始文本按原样返回。.
//
// The key is a Chinese constant in the source/从文字中提取字节并创建它们。(scripts 助理). 用手
// 因为我没有打字 writeErr 传入的运行时值和键不会错位，并且会被悄悄错过。.
//
// 格式化字符串常量(%d·%s·%w 等等)因为格式化结果与key不同，, 在呼叫部门 fmt.Sprintf/
// fmt.Errorf 在移交给之前 trMsg(模板) 首先将模板本身替换为.

// trMsg 将原始中文消息替换为活动输出语言。. 不在目录或活动语言中
// 如果是中文字符，则原样返回原文。. config.ActiveLanguage 运行时设置(设置表)到
// 因为它反映了, 用户选择的语言立即应用.
func trMsg(source string) string {
	if m, ok := errCatalog[source]; ok {
		if v, ok2 := m[config.ActiveLanguage()]; ok2 && v != "" {
			return v
		}
	}
	return source
}

// logT translates an operator-facing console log template (config.T) so server
// log.Printf sites can wrap a template without each file importing config.
func logT(source string) string { return config.T(source) }

// loadOutputLanguage resolves the user-facing output language from the source of
// truth: the runtime setting (settings table) over config.Language() (ARGUS_LANG >
// config file > ko). Call it only to (re)fill the cache; it does a DB query and a
// config file read.
func (s *Server) loadOutputLanguage() string {
	if s.m != nil && s.m.pg != nil {
		if v, ok, _ := s.m.pg.GetSetting(settingOutputLanguage); ok {
			if nv := config.NormalizeLanguage(v); nv != "" {
				return nv
			}
		}
	}
	return config.Language()
}

// outputLanguage returns the cached output language, resolving it on first use.
func (s *Server) outputLanguage() string {
	if v, ok := s.outputLang.Load().(string); ok && v != "" {
		return v
	}
	v := s.loadOutputLanguage()
	s.outputLang.Store(v)
	return v
}
