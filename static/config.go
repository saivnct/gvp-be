package static

const (
	ServerName                      = "gvpServer"
	ServerEventName                 = "gvpServerEvent"
	MaxResendAuthenCode             = 3
	TimeResetEmailCreateAccountLock = 60     //60 mins
	AuthencodeTimeOut               = 5 * 60 //5 min
	MaxNumberAuthencodeFail         = 3
	JWTTTL                          = 60 //60 mins
	//MaxFileSize = 1 << 10 //1 KB
	MaxImageFileSize        = 20 * 1024 << 10   //20 MB
	MaxMediaPreviewFileSize = 100 * 1024 << 10  //100 MB
	MaxMediaFileSize        = 2048 * 1024 << 10 //2048 MB = 2GB
	S3NamePrefxix           = "s3_gvp_"

	Pagination_Default_PageSize = 10
	Pagination_Max_PageSize     = 50

	NewsID_Avatar = "Avatar"
)

func JWTKey() []byte {
	return []byte("GVP JWT KEY")
}
