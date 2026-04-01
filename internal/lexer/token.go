package lexer

type TokenType int

const (
	TokenIdent  TokenType = iota
	TokenString
	TokenNumber

	TokenSelect
	TokenFrom
	TokenWhere
	TokenAnd
	TokenOr
	TokenNot
	TokenIs
	TokenNull
	TokenLike
	TokenIn
	TokenOrder
	TokenBy
	TokenAsc
	TokenDesc
	TokenLimit
	TokenAs
	TokenUpdate
	TokenSet
	TokenDelete
	TokenInsert
	TokenInto
	TokenValues
	TokenTrue
	TokenFalse

	TokenStar
	TokenComma
	TokenLParen
	TokenRParen
	TokenEq
	TokenNeq
	TokenLt
	TokenGt
	TokenLte
	TokenGte
	TokenSemicolon

	TokenEOF
	TokenIllegal
)

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

var keywords = map[string]TokenType{
	"SELECT": TokenSelect,
	"FROM":   TokenFrom,
	"WHERE":  TokenWhere,
	"AND":    TokenAnd,
	"OR":     TokenOr,
	"NOT":    TokenNot,
	"IS":     TokenIs,
	"NULL":   TokenNull,
	"LIKE":   TokenLike,
	"IN":     TokenIn,
	"ORDER":  TokenOrder,
	"BY":     TokenBy,
	"ASC":    TokenAsc,
	"DESC":   TokenDesc,
	"LIMIT":  TokenLimit,
	"AS":     TokenAs,
	"UPDATE": TokenUpdate,
	"SET":    TokenSet,
	"DELETE": TokenDelete,
	"INSERT": TokenInsert,
	"INTO":   TokenInto,
	"VALUES": TokenValues,
	"TRUE":   TokenTrue,
	"FALSE":  TokenFalse,
}

func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}
