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

// 帐变类型
const (
	TransactionIn                    = 151 //场馆转入
	TransactionOut                   = 152 //场馆转出
	TransactionInFail                = 153 //场馆转入失败补回
	TransactionOutFail               = 154 //场馆转出失败扣除
	TransactionDeposit               = 155 //存款
	TransactionWithDraw              = 156 //提现
	TransactionUpPoint               = 157 //后台上分
	TransactionDownPoint             = 158 //后台下分
	TransactionDownPointBack         = 159 //后台下分回退
	TransactionDividend              = 160 //中心钱包红利派发
	TransactionFinanceDownPoint      = 162 //财务下分
	TransactionWithDrawFail          = 163 //提现失败
	TransactionValetDeposit          = 164 //代客充值
	TransactionValetWithdraw         = 165 //代客提款
	TransactionAgencyDeposit         = 166 //代理充值
	TransactionAgencyWithdraw        = 167 //代理提款
	TransactionPlatUpPoint           = 168 //后台场馆上分
	TransactionPlatDividend          = 169 //场馆红利派发
	TransactionFirstDepositDividend  = 171 //首存活动红利
	TransactionInviteDividend        = 172 //邀请好友红利
	TransactionBet                   = 173 //投注
	TransactionBetCancel             = 174 //投注取消
	TransactionPayout                = 175 //派彩
	TransactionResettlePlus          = 176 //重新结算加币
	TransactionResettleDeduction     = 177 //重新结算减币
	TransactionCancelPayout          = 178 //取消派彩
	TransactionPromoPayout           = 179 //场馆活动派彩
	TransactionEBetTCPrize           = 600 //EBet宝箱奖金
	TransactionEBetLimitRp           = 601 //EBet限量红包
	TransactionEBetLuckyRp           = 602 //EBet幸运红包
	TransactionEBetMasterPayout      = 603 //EBet大赛派彩
	TransactionEBetMasterRegFee      = 604 //EBet大赛报名费
	TransactionEBetBetPrize          = 605 //EBet投注奖励
	TransactionEBetReward            = 606 //EBet打赏
	TransactionEBetMasterPrizeDeduct = 607 //EBet大赛奖金取回
	TransactionWMReward              = 608 //WM打赏
	TransactionSBODividend           = 609 //SBO红利
	TransactionSBOReward             = 610 //SBO打赏
	TransactionSBOBuyLiveCoin        = 611 //SBO 购买LiveCoin
	TransactionSignDividend          = 612 //天天签到活动红利
	TransactionCQ9Dividend           = 613 //CQ9游戏红利
	TransactionCQ9PromoPayout        = 614 //CQ9活动派彩
	TransactionPlayStarPrize         = 615 //Playstar积宝奖金
	TransactionSpadeGamingRp         = 616 //SpadeGaming红包
	TransactionAEReward              = 617 //AE打赏
	TransactionAECancelReward        = 618 //AE取消打赏
	TransactionOfflineDeposit        = 619 //线下转卡存款
	TransactionUSDTOfflineDeposit    = 620 //USDT线下存款
	TransactionEVOPrize              = 621 //游戏奖金(EVO)
	TransactionEVOPromote            = 622 //推广(EVO)
	TransactionEVOJackpot            = 623 //头奖(EVO)
	TransactionCommissionDraw        = 624 //佣金提取
)

const (
	COTransactionSubRebate  = 170 //下级返水
	COTransactionRebate     = 161 //会员返水
	COTransactionReceive    = 751 //佣金发放
	COTransactionDraw       = 752 //佣金提取
	COTransactionRation     = 753 //佣金下发
	COTransactionDrawBack   = 754 //佣金提取补回
	COTransactionRationBack = 755 //佣金下发补回
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
