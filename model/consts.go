package model

var devices = map[int]uint8{
	24: 1,
	25: 2,
	35: 4,
	36: 3,
}

const (
	TyDevice         = 1 //设备号
	TyIP             = 2 //ip地址
	TyEmail          = 3 //邮箱地址
	TyPhone          = 4 //电话号码
	TyBankcard       = 5 //银行卡
	TyVirtualAccount = 6 //虚拟币地址
)

var (
	Devices = map[int]bool{
		DeviceTypeWeb:            true,
		DeviceTypeH5:             true,
		DeviceTypeAndroidFlutter: true,
		DeviceTypeIOSFlutter:     true,
	}
	WebDevices = map[int]bool{
		DeviceTypeWeb: true,
		DeviceTypeH5:  true,
	}
	AppDevices = map[int]bool{
		DeviceTypeAndroidFlutter: true,
		DeviceTypeIOSFlutter:     true,
	}
)

// 设备端类型
const (
	DeviceTypeWeb            = 24 //web
	DeviceTypeH5             = 25 //h5
	DeviceTypeAndroidFlutter = 35 //android_flutter
	DeviceTypeIOSFlutter     = 36 //ios_flutter
)

// 场馆转账类型
const (
	TransferIn           = 181 //场馆转入
	TransferOut          = 182 //场馆转出
	TransferUpPoint      = 183 //后台场馆上分
	TransferResetBalance = 184 //场馆钱包清零
	TransferDividend     = 185 //场馆红利
)

// 红利审核状态
const (
	DividendReviewing    = 231 //红利审核中
	DividendReviewPass   = 232 //红利审核通过
	DividendReviewReject = 233 //红利审核不通过
)

// 红利发放状态
const (
	DividendFailed      = 236 //红利发放失败
	DividendSuccess     = 237 //红利发放成功
	DividendPlatDealing = 238 //红利发放场馆处理中
)

// 后台调整类型
const (
	AdjustUpMode   = 251 // 上分
	AdjustDownMode = 252 // 下分
)

// 后台上下分审核状态
const (
	AdjustReviewing    = 256 //后台调整审核中
	AdjustReviewPass   = 257 //后台调整审核通过
	AdjustReviewReject = 258 //后台调整审核不通过
)

const (
	RecordTradeDeposit  int = 271 // 存款
	RecordTradeWithdraw int = 272 // 取款
	RecordTradeTransfer int = 273 // 转账
	RecordTradeDividend int = 274 // 红利
	RecordTradeRebate   int = 275 // 返水/佣金
	RecordTradeAdjust   int = 278 // 调整
)

// 返水审核状态
const (
	RebateReviewing    = 291 //返水审核中
	RebateReviewPass   = 292 //返水已发放
	RebateReviewReject = 293 //返水已拒绝
)

// 取款状态
const (
	WithdrawReviewing     = 371 //审核中
	WithdrawReviewReject  = 372 //审核拒绝
	WithdrawDealing       = 373 //出款中
	WithdrawSuccess       = 374 //提款成功
	WithdrawFailed        = 375 //出款失败
	WithdrawAbnormal      = 376 //异常订单
	WithdrawAutoPayFailed = 377 // 代付失败
	WithdrawHangup        = 378 // 挂起
	WithdrawDispatched    = 379 // 已派单
)

// 发送邮件
const (
	EmailModifyPassword = 1
	EmailForgetPassword = 2
)
