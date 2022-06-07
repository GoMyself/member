package model

import "database/sql"

type Member struct {
	UID                 string `db:"uid" json:"uid" redis:"uid"`
	Username            string `db:"username" json:"username" redis:"username"`                                     //会员名
	Password            string `db:"password" json:"password" redis:"password"`                                     //密码
	Birth               string `db:"birth" json:"birth" redis:"birth"`                                              //生日日期
	BirthHash           string `db:"birth_hash" json:"birth_hash" redis:"birth_hash"`                               //生日日期哈希
	RealnameHash        string `db:"realname_hash" json:"realname_hash" redis:"realname_hash"`                      //真实姓名哈希
	EmailHash           string `db:"email_hash" json:"email_hash" redis:"email_hash"`                               //邮件地址哈希
	PhoneHash           string `db:"phone_hash" json:"phone_hash" redis:"phone_hash"`                               //电话号码哈希
	ZaloHash            string `db:"zalo_hash" json:"zalo_hash" redis:"zalo_hash"`                                  //zalo哈希
	Prefix              string `db:"prefix" json:"prefix" redis:"prefix"`                                           //站点前缀
	Tester              string `db:"tester" json:"tester" redis:"tester"`                                           //1正式 0测试
	WithdrawPwd         uint64 `db:"withdraw_pwd" json:"withdraw_pwd" redis:"withdraw_pwd"`                         //取款密码哈希
	Regip               string `db:"regip" json:"regip" redis:"regip"`                                              //注册IP
	RegDevice           string `db:"reg_device" json:"reg_device" redis:"reg_device"`                               //注册设备号
	RegUrl              string `db:"reg_url" json:"reg_url" redis:"reg_url"`                                        //注册链接
	CreatedAt           uint32 `db:"created_at" json:"created_at" redis:"created_at"`                               //注册时间
	LastLoginIp         string `db:"last_login_ip" json:"last_login_ip" redis:"last_login_ip"`                      //最后登陆ip
	LastLoginAt         uint32 `db:"last_login_at" json:"last_login_at" redis:"last_login_at"`                      //最后登陆时间
	SourceId            uint8  `db:"source_id" json:"source_id" redis:"source_id"`                                  //注册来源 1 pc 2h5 3 app
	FirstDepositAt      uint32 `db:"first_deposit_at" json:"first_deposit_at" redis:"first_deposit_at"`             //首充时间
	FirstDepositAmount  string `db:"first_deposit_amount" json:"first_deposit_amount" redis:"first_deposit_amount"` //首充金额
	FirstBetAt          uint32 `db:"first_bet_at" json:"first_bet_at" redis:"first_bet_at"`                         //首投时间
	FirstBetAmount      string `db:"first_bet_amount" json:"first_bet_amount" redis:"first_bet_amount"`             //首投金额
	SecondDepositAt     uint32 `db:"second_deposit_at" json:"second_deposit_at"`                                    //二存时间
	SecondDepositAmount string `db:"second_deposit_amount" json:"second_deposit_amount"`                            //二充金额
	TopUid              string `db:"top_uid" json:"top_uid" redis:"top_uid"`                                        //总代uid
	TopName             string `db:"top_name" json:"top_name" redis:"top_name"`                                     //总代代理
	ParentUid           string `db:"parent_uid" json:"parent_uid" redis:"parent_uid"`                               //上级uid
	ParentName          string `db:"parent_name" json:"parent_name" redis:"parent_name"`                            //上级代理
	BankcardTotal       uint8  `db:"bankcard_total" json:"bankcard_total" redis:"bankcard_total"`                   //用户绑定银行卡的数量
	LastLoginDevice     string `db:"last_login_device" json:"last_login_device" redis:"last_login_device"`          //最后登陆设备
	LastLoginSource     int    `db:"last_login_source" json:"last_login_source" redis:"last_login_source"`          //上次登录设备来源:1=pc,2=h5,3=ios,4=andriod
	Remarks             string `db:"remarks" json:"remarks" redis:"remarks"`                                        //备注
	State               uint8  `db:"state" json:"state" redis:"state"`                                              //状态 1正常 2禁用
	Level               int    `db:"level" json:"level" redis:"level" redis:"level"`                                //等级
	Balance             string `db:"balance" json:"balance" redis:"balance"`                                        //余额
	LockAmount          string `db:"lock_amount" json:"lock_amount" redis:"lock_amount"`                            //锁定金额
	Commission          string `db:"commission" json:"commission" redis:"commission"`                               //佣金
	GroupName           string `db:"group_name" json:"group_name" redis:"group_name"`                               //团队名称
	AgencyType          int64  `db:"agency_type" json:"agency_type" redis:"agency_type"`                            //391团队代理 393普通代理
	Address             string `db:"address" json:"address" redis:"address"`                                        //收货地址
	Avatar              string `db:"avatar" json:"avatar" redis:"avatar"`                                           //收货地址
}

type MemberMaxRebate struct {
	ZR               sql.NullFloat64 `db:"zr" json:"zr"`                                 //真人返水
	QP               sql.NullFloat64 `db:"qp" json:"qp"`                                 //棋牌返水
	TY               sql.NullFloat64 `db:"ty" json:"ty"`                                 //体育返水
	DJ               sql.NullFloat64 `db:"dj" json:"dj"`                                 //电竞返水
	DZ               sql.NullFloat64 `db:"dz" json:"dz"`                                 //电游返水
	CP               sql.NullFloat64 `db:"cp" json:"cp"`                                 //彩票返水
	FC               sql.NullFloat64 `db:"fc" json:"fc"`                                 //斗鸡返水
	BY               sql.NullFloat64 `db:"by" json:"by"`                                 //捕鱼返水
	CgHighRebate     sql.NullFloat64 `db:"cg_high_rebate" json:"cg_high_rebate"`         //高频彩返点
	CgOfficialRebate sql.NullFloat64 `db:"cg_official_rebate" json:"cg_official_rebate"` //官方彩返点
}

type MemberData struct {
	Member
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	IsRisk   int    `json:"is_risk"`
}

type MBBalance struct {
	Balance    string `db:"balance" json:"balance"`         //余额
	LockAmount string `db:"lock_amount" json:"lock_amount"` //锁定额度
}

type MemberInfos struct {
	UID           string `db:"uid" json:"uid" redis:"uid"`
	Username      string `db:"username" json:"username" redis:"username"`
	RealnameHash  string `db:"realname_hash" json:"realname_hash" redis:"realname_hash"`
	EmailHash     string `db:"email_hash" json:"email_hash" redis:"email_hash"`
	PhoneHash     string `db:"phone_hash" json:"phone_hash" redis:"phone_hash"`
	ZaloHash      string `db:"zalo_hash" json:"zalo_hash" redis:"zalo_hash"`
	Address       string `db:"address" json:"address" redis:"address"`
	CreatedAt     uint32 `db:"created_at" json:"created_at" redis:"created_at"`
	BankcardTotal int    `db:"bankcard_total" json:"bankcard_total" redis:"bankcard_total"`
	Level         int    `db:"level" json:"level" redis:"level"`
}

type MemberInfosData struct {
	MemberInfos
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Zalo     string `json:"zalo"`
}

type BankCard struct {
	ID           string `db:"id" json:"id"`
	UID          string `db:"uid" json:"uid"`
	Username     string `db:"username" json:"username"`
	BankAddress  string `db:"bank_address" json:"bank_address"`
	BankID       string `db:"bank_id" json:"bank_id"`
	BankBranch   string `db:"bank_branch_name" json:"bank_branch_name"`
	State        int    `db:"state" json:"state"`
	BankcardHash string `db:"bank_card_hash" json:"bank_card_hash"`
	CreatedAt    string `db:"created_at" json:"created_at"`
}

type BankcardData struct {
	BankCard
	RealName string `json:"realname" name:"realname"`
	Bankcard string `json:"bank_card" name:"bankcard"`
}

type MemberRebate struct {
	UID              string `db:"uid" json:"uid"`
	ParentUID        string `db:"parent_uid" json:"parent_uid"`
	ZR               string `db:"zr" json:"zr" redis:"zr"`                                                 //真人返水
	QP               string `db:"qp" json:"qp" redis:"qp"`                                                 //棋牌返水
	TY               string `db:"ty" json:"ty" redis:"ty"`                                                 //体育返水
	DJ               string `db:"dj" json:"dj" redis:"dj"`                                                 //电竞返水
	DZ               string `db:"dz" json:"dz" redis:"dz"`                                                 //电游返水
	CP               string `db:"cp" json:"cp" redis:"cp"`                                                 //彩票返水
	FC               string `db:"fc" json:"fc" redis:"fc"`                                                 //斗鸡返水
	BY               string `db:"by" json:"by" redis:"by"`                                                 //捕鱼返水
	CGHighRebate     string `db:"cg_high_rebate" json:"cg_high_rebate" redis:"cg_high_rebate"`             //CG高频彩返点
	CGOfficialRebate string `db:"cg_official_rebate" json:"cg_official_rebate" redis:"cg_official_rebate"` //CG官方彩返点
	CreatedAt        uint32 `db:"created_at" json:"created_at"`
	Prefix           string `db:"prefix" json:"prefix"`
}

// 场馆转账数据
type TransferData struct {
	T   int64             `json:"t"`
	D   []Transfer        `json:"d"`
	Agg map[string]string `json:"agg"`
}

//转账记录
type Transfer struct {
	ID           string  `json:"id"`
	UID          string  `json:"uid"`
	BillNo       string  `json:"bill_no"`
	PlatformId   string  `json:"platform_id"`
	Username     string  `json:"username"`
	TransferType int     `json:"transfer_type"`
	Amount       float64 `json:"amount"`
	BeforeAmount float64 `json:"before_amount"`
	AfterAmount  float64 `json:"after_amount"`
	CreatedAt    uint64  `json:"created_at"`
	State        int     `json:"state"`
	Automatic    int     `json:"automatic"`
	ConfirmName  string  `json:"confirm_name"`
}

// 游戏记录数据
type GameRecordData struct {
	T   int64             `json:"t"`
	D   []GameRecord      `json:"d"`
	Agg map[string]string `json:"agg"`
}

//游戏投注记录结构
type GameRecord struct {
	ID             string  `db:"column:id" json:"id" form:"id"`
	RowId          string  `db:"column:row_id" json:"row_id" form:"row_id"`
	BillNo         string  `db:"column:bill_no" json:"bill_no" form:"bill_no"`
	ApiType        int     `db:"column:api_type" json:"api_type" form:"api_type"`
	ApiTypes       string  `json:"api_types"`
	PlayerName     string  `db:"column:player_name" json:"player_name" form:"player_name"`
	Name           string  `db:"column:name" json:"name" form:"name"`
	Uid            string  `db:"column:uid" json:"uid" form:"uid"`
	NetAmount      float64 `db:"column:net_amount" json:"net_amount" form:"net_amount"`
	BetTime        int64   `db:"column:bet_time" json:"bet_time" form:"bet_time"`
	StartTime      int64   `db:"column:start_time" json:"start_time" form:"start_time"`
	Resettle       uint8   `db:"column:resettle" json:"resettle" form:"resettle"`
	Presettle      uint8   `db:"column:presettle" json:"presettle" form:"presettle"`
	GameType       string  `db:"column:game_type" json:"game_type" form:"game_type"`
	BetAmount      float64 `db:"column:bet_amount" json:"bet_amount" form:"bet_amount"`
	ValidBetAmount float64 `db:"column:valid_bet_amount" json:"valid_bet_amount" form:"valid_bet_amount"`
	Flag           int     `db:"column:flag" json:"flag" form:"flag"`
	PlayType       string  `db:"column:play_type" json:"play_type" form:"play_type"`
	CopyFlag       int     `db:"column:copy_flag" json:"copy_flag" form:"copy_flag"`
	FilePath       string  `db:"column:file_path" json:"file_path" form:"file_path"`
	Prefix         string  `db:"column:prefix" json:"prefix" form:"prefix"`
	Result         string  `db:"column:result" json:"result" form:"result"`
	CreatedAt      uint64  `db:"column:created_at" json:"created_at" form:"created_at"`
	UpdatedAt      uint64  `db:"column:updated_at" json:"updated_at" form:"updated_at"`
	ApiName        string  `db:"column:api_name" json:"api_name" form:"api_name"`
	ApiBillNo      string  `db:"column:api_bill_no" json:"api_bill_no" form:"api_bill_no"`
	MainBillNo     string  `db:"column:main_bill_no" json:"main_bill_no" form:"main_bill_no"`
	IsUse          int     `db:"column:is_use" json:"is_use" form:"is_use"`
	FlowQuota      int64   `db:"column:flow_quota" json:"flow_quota" form:"flow_quota"`
	GameName       string  `db:"column:game_name" json:"game_name" form:"game_name"`
	HandicapType   string  `db:"column:handicap_type" json:"handicap_type" form:"handicap_type"`
	Handicap       string  `db:"column:handicap" json:"handicap" form:"handicap"`
	Odds           float64 `db:"column:odds" json:"odds" form:"odds"`
	BallType       int     `db:"column:ball_type" json:"ball_type" form:"ball_type"`
	SettleTime     int64   `db:"column:settle_time" json:"settle_time" form:"settle_time"`
	ApiBetTime     uint64  `db:"column:api_bet_time" json:"api_bet_time" form:"api_bet_time"`
	ApiSettleTime  uint64  `db:"column:api_settle_time" json:"api_settle_time" form:"api_settle_time"`
	AgencyUid      string  `db:"column:agency_uid" json:"agency_uid" form:"agency_uid"`
	AgencyName     string  `db:"column:agency_name" json:"agency_name" form:"agency_name"`
	AgencyGid      string  `db:"column:agency_gid" json:"agency_gid" form:"agency_gid"` //代理团队id
	IsRisk         int     `db:"-" json:"is_risk"`
}

//账变表
type MemberTransaction struct {
	AfterAmount  string `db:"after_amount"`  //账变后的金额
	Amount       string `db:"amount"`        //用户填写的转换金额
	BeforeAmount string `db:"before_amount"` //账变前的金额
	BillNo       string `db:"bill_no"`       //转账|充值|提现ID
	CashType     int    `db:"cash_type"`     //0:转入1:转出2:转入失败补回3:转出失败扣除4:存款5:提现
	CreatedAt    int64  `db:"created_at"`    //
	ID           string `db:"id"`            //
	UID          string `db:"uid"`           //用户ID
	Username     string `db:"username"`      //用户名
	Prefix       string `json:"prefix" db:"prefix"`
}

// 帐变数据
type TransactionData struct {
	T   int64             `json:"t"`
	D   []Transaction     `json:"d"`
	Agg map[string]string `json:"agg"`
}

type Transaction struct {
	ID           string  `db:"column:id" json:"id" form:"id"`
	BillNo       string  `db:"column:bill_no" json:"bill_no" form:"bill_no"`
	Uid          string  `db:"column:uid" json:"uid" form:"uid"`
	Username     string  `db:"column:username" json:"username" form:"username"`
	CashType     int     `db:"column:cash_type" json:"cash_type" form:"cash_type"`
	Amount       float64 `db:"column:amount" json:"amount" form:"amount"`
	BeforeAmount float64 `db:"column:before_amount" json:"before_amount" form:"before_amount"`
	AfterAmount  float64 `db:"column:after_amount" json:"after_amount" form:"after_amount"`
	CreatedAt    uint64  `db:"column:created_at" json:"created_at" form:"created_at"`
}

// 取款数据
type WithdrawData struct {
	T   int64             `json:"t"`
	D   []Withdraw        `json:"d"`
	Agg map[string]string `json:"agg"`
}

// Withdraw 出款
type Withdraw struct {
	ID                string  `db:"id" json:"id"`                                   //
	Prefix            string  `db:"prefix" json:"prefix"`                           //转账后的金额
	BID               string  `db:"bid" json:"bid"`                                 //转账前的金额
	Flag              int     `db:"flag" json:"flag"`                               //
	OID               string  `db:"oid" json:"oid"`                                 //转账前的金额
	UID               string  `db:"uid" json:"uid"`                                 //用户ID
	Username          string  `db:"username" json:"username"`                       //用户名
	TopUid            string  `db:"top_uid" json:"top_uid"`                         //总代uid
	TopName           string  `db:"top_name" json:"top_name"`                       //总代代理
	ParentUid         string  `db:"parent_uid" json:"parent_uid"`                   //上级uid
	ParentName        string  `db:"parent_name" json:"parent_name"`                 //上级代理
	PID               string  `db:"pid" json:"pid"`                                 //用户ID
	Amount            float64 `db:"amount" json:"amount"`                           //金额
	State             int     `db:"state" json:"state"`                             //0:待确认:1存款成功2:已取消
	Automatic         int     `db:"automatic" json:"automatic"`                     //1:自动转账2:脚本确认3:人工确认
	BankName          string  `db:"bank_name" json:"bank_name"`                     //银行名
	RealName          string  `db:"real_name" json:"real_name"`                     //持卡人姓名
	CardNO            string  `db:"card_no" json:"card_no"`                         //银行卡号
	CreatedAt         int64   `db:"created_at" json:"created_at"`                   //
	ConfirmAt         int64   `db:"confirm_at" json:"confirm_at"`                   //确认时间
	ConfirmUID        string  `db:"confirm_uid" json:"confirm_uid"`                 //确认人uid
	ConfirmName       string  `db:"confirm_name" json:"confirm_name"`               //确认人名
	ReviewRemark      string  `db:"review_remark" json:"review_remark"`             //确认人名
	WithdrawAt        int64   `db:"withdraw_at" json:"withdraw_at"`                 //三方场馆ID
	WithdrawRemark    string  `db:"withdraw_remark" json:"withdraw_remark"`         //确认人名
	WithdrawUID       string  `db:"withdraw_uid" json:"withdraw_uid"`               //确认人uid
	WithdrawName      string  `db:"withdraw_name" json:"withdraw_name"`             //确认人名
	FinanceType       int     `db:"finance_type" json:"finance_type"`               // 财务类型 442=提款 444=代客提款 446=代理提款
	LastDepositAmount float64 `db:"last_deposit_amount" json:"last_deposit_amount"` // 上笔成功存款金额
	RealNameHash      string  `db:"real_name_hash" json:"real_name_hash"`           //真实姓名哈希
	HangUpUID         string  `db:"hang_up_uid" json:"hang_up_uid"`                 // 挂起人uid
	HangUpRemark      string  `db:"hang_up_remark" json:"hang_up_remark"`           // 挂起备注
	HangUpName        string  `db:"hang_up_name" json:"hang_up_name"`               //  挂起人名字
	RemarkID          int     `db:"remark_id" json:"remark_id"`                     // 挂起原因ID
	HangUpAt          int     `db:"hang_up_at" json:"hang_up_at"`                   //  挂起时间
	ReceiveAt         int64   `db:"receive_at" json:"receive_at"`                   //领取时间
	WalletFlag        int     `db:"wallet_flag" json:"wallet_flag"`                 //钱包类型:1=中心钱包,2=佣金钱包
}

// 存款数据
type DepositData struct {
	T   int64             `json:"t"`
	D   []Deposit         `json:"d"`
	Agg map[string]string `json:"agg"`
}

// Deposit 存款
type Deposit struct {
	ID              string  `db:"id" json:"id"`                               //
	Prefix          string  `db:"prefix" json:"prefix"`                       //转账后的金额
	OID             string  `db:"oid" json:"oid"`                             //转账前的金额
	UID             string  `db:"uid" json:"uid"`                             //用户ID
	Username        string  `db:"username" json:"username"`                   //用户名
	TopUid          string  `db:"top_uid" json:"top_uid"`                     //总代uid
	TopName         string  `db:"top_name" json:"top_name"`                   //总代代理
	ParentUid       string  `db:"parent_uid" json:"parent_uid"`               //上级uid
	ParentName      string  `db:"parent_name" json:"parent_name"`             //上级代理
	ChannelID       string  `db:"channel_id" json:"channel_id"`               //
	CID             string  `db:"cid" json:"cid"`                             //分类ID
	PID             string  `db:"pid" json:"pid"`                             //用户ID
	FinanceType     int     `db:"finance_type" json:"finance_type"`           //
	Amount          float64 `db:"amount" json:"amount"`                       //金额
	USDTFinalAmount float64 `db:"usdt_final_amount" json:"usdt_final_amount"` // 到账金额
	USDTApplyAmount float64 `db:"usdt_apply_amount" json:"usdt_apply_amount"` // 提单金额
	Rate            float64 `db:"rate" json:"rate"`                           // 汇率
	State           int     `db:"state" json:"state"`                         //0:待确认:1存款成功2:已取消
	Automatic       int     `db:"automatic" json:"automatic"`                 //1:自动转账2:脚本确认3:人工确认
	CreatedAt       int64   `db:"created_at" json:"created_at"`               //
	CreatedUID      string  `db:"created_uid" json:"created_uid"`             //三方场馆ID
	CreatedName     string  `db:"created_name" json:"created_name"`           //确认人名
	ReviewRemark    string  `db:"review_remark" json:"review_remark"`         //确认人名
	ConfirmAt       int64   `db:"confirm_at" json:"confirm_at"`               //确认时间
	ConfirmUID      string  `db:"confirm_uid" json:"confirm_uid"`             //确认人uid
	ConfirmName     string  `db:"confirm_name" json:"confirm_name"`           //确认人名
	IsRisk          int     `db:"-" json:"is_risk"`                           //是否风控
	ProtocolType    string  `db:"protocol_type" json:"protocol_type"`         //地址类型 trc20 erc20
	Address         string  `db:"address" json:"address"`                     //收款地址
	HashID          string  `db:"hash_id" json:"hash_id"`                     //区块链订单号
	Flag            int     `db:"flag" json:"flag"`                           // 1 三方订单 2 三方usdt订单 3 线下转卡订单 4 线下转usdt订单
	BankcardID      string  `db:"bankcard_id" json:"bankcard_id"`             // 线下转卡 收款银行卡id
	ManualRemark    string  `db:"manual_remark" json:"manual_remark"`         // 线下转卡订单附言
	BankCode        string  `db:"bank_code" json:"bank_code"`                 // 银行编号
	BankNo          string  `db:"bank_no" json:"bank_no"`                     // 银行卡号
}

// 红利数据
type DividendData struct {
	T   int64             `json:"t"`
	D   []Dividend        `json:"d"`
	Agg map[string]string `json:"agg"`
}

type Dividend struct {
	ID            string  `db:"id" json:"id"`
	UID           string  `db:"uid" json:"uid"`
	Username      string  `db:"username" json:"username"`
	TopUid        string  `db:"top_uid" json:"top_uid"`         //总代uid
	TopName       string  `db:"top_name" json:"top_name"`       //总代代理
	ParentUid     string  `db:"parent_uid" json:"parent_uid"`   //上级uid
	ParentName    string  `db:"parent_name" json:"parent_name"` //上级代理
	Prefix        string  `db:"prefix" json:"prefix"`
	Wallet        int     `db:"wallet" json:"wallet"`
	Ty            int     `db:"ty" json:"ty"`
	WaterLimit    uint8   `db:"water_limit" json:"water_limit"`
	PlatformID    string  `db:"platform_id" json:"platform_id"`
	Amount        float64 `db:"amount" json:"amount"`
	HandOutAmount float64 `db:"hand_out_amount" json:"hand_out_amount"`
	WaterFlow     float64 `db:"water_flow" json:"water_flow"`
	Notify        uint8   `db:"notify" json:"notify"`
	State         int     `db:"state" json:"state"`
	HandOutState  int     `db:"hand_out_state" json:"hand_out_state"`
	Automatic     int     `db:"automatic" json:"automatic"`
	Remark        string  `db:"remark" json:"remark"`
	ReviewRemark  string  `db:"review_remark" json:"review_remark"`
	ApplyAt       uint64  `db:"apply_at" json:"apply_at"`
	ApplyUid      string  `db:"apply_uid" json:"apply_uid"`
	ApplyName     string  `db:"apply_name" json:"apply_name"`
	ReviewAt      uint64  `db:"review_at" json:"review_at"`
	ReviewUid     string  `db:"review_uid" json:"review_uid"`
	ReviewName    string  `db:"review_name" json:"review_name"`
	IsRisk        int     `db:"-" json:"is_risk"`
}

// 返水数据
//type RebateData struct {
//	T   int64             `json:"t"`
//	D   []Rebate          `json:"d"`
//	Agg map[string]string `json:"agg"`
//}

//type Rebate struct {
//	ID           string  `db:"id" json:"id"`
//	UID          string  `db:"uid" json:"uid"`
//	Username     string  `db:"username" json:"username"`
//	TopUid       string  `db:"top_uid" json:"top_uid"`         //总代uid
//	TopName      string  `db:"top_name" json:"top_name"`       //总代代理
//	ParentUid    string  `db:"parent_uid" json:"parent_uid"`   //上级uid
//	ParentName   string  `db:"parent_name" json:"parent_name"` //上级代理
//	RebateAt     uint64  `db:"rebate_at" json:"rebate_at"`
//	RationAt     uint64  `db:"ration_at" json:"ration_at"`
//	ShouldAmount float64 `db:"should_amount" json:"should_amount"`
//	RebateAmount float64 `db:"rebate_amount" json:"rebate_amount"`
//	CheckAt      uint64  `db:"check_at" json:"check_at"`
//	State        int     `db:"state" json:"state"`
//	CheckNote    string  `db:"check_note" json:"check_note"`
//	RationFlag   int     `db:"ration_flag" json:"ration_flag"`
//	CheckUID     string  `db:"check_uid" json:"check_uid"`
//	CheckName    string  `db:"check_name" json:"check_name"`
//	CreateAt     uint64  `db:"create_at" json:"create_at"`
//	IsRisk       int     `db:"-" json:"is_risk"`
//}

// MemberAdjust db structure
type MemberAdjust struct {
	ID            string  `db:"id" json:"id"`
	UID           string  `db:"uid" json:"uid"` // 会员id
	Prefix        string  `db:"prefix" json:"prefix"`
	Ty            int     `db:"ty" json:"ty"`                         // 来源
	Username      string  `db:"username" json:"username"`             // 会员名
	TopUid        string  `db:"top_uid" json:"top_uid"`               // 总代uid
	TopName       string  `db:"top_name" json:"top_name"`             // 总代代理
	ParentUid     string  `db:"parent_uid" json:"parent_uid"`         // 上级uid
	ParentName    string  `db:"parent_name" json:"parent_name"`       // 上级代理
	Amount        float64 `db:"amount" json:"amount"`                 // 调整金额
	AdjustType    int     `db:"adjust_type" json:"adjust_type"`       // 调整类型:1=系统调整,2=输赢调整,3=线下转卡充值
	AdjustMode    int64   `db:"adjust_mode" json:"adjust_mode"`       // 调整方式:1=上分,2=下分
	IsTurnover    int64   `db:"is_turnover" json:"is_turnover"`       // 是否需要流水限制:1=需要,0=不需要
	TurnoverMulti int64   `db:"turnover_multi" json:"turnover_multi"` // 流水倍数
	ApplyRemark   string  `db:"apply_remark" json:"apply_remark"`     // 申请备注
	ReviewRemark  string  `db:"review_remark" json:"review_remark"`   // 审核备注
	State         int     `db:"state" json:"state"`                   // 状态:1=审核中,2=审核通过,3=审核未通过
	HandOutState  int64   `db:"hand_out_state" json:"hand_out_state"` // 上下分状态 1 失败 2成功 3场馆上分处理中
	Images        string  `db:"images" json:"images"`                 // 图片地址
	ApplyAt       int64   `db:"apply_at" json:"apply_at"`             // 申请时间
	ApplyUid      string  `db:"apply_uid" json:"apply_uid"`           // 申请人uid
	ApplyName     string  `db:"apply_name" json:"apply_name"`         // 申请人
	ReviewAt      int64   `db:"review_at" json:"review_at"`           // 审核时间
	ReviewUid     string  `db:"review_uid" json:"review_uid"`         // 审核人uid
	ReviewName    string  `db:"review_name" json:"review_name"`       // 审核人
	IsRisk        int     `db:"-" json:"is_risk"`
}

// MemberListData 会员列表
type MemberListData struct {
	T         int             `json:"t"`
	S         int             `json:"s"`
	EnableMod bool            `json:"enable_mod"`
	D         []MemberListCol `json:"d"`
	Agg       MemberAggData   `json:"agg"`
}

type MemberListCol struct {
	UID              string  `json:"uid" db:"uid"`
	Username         string  `json:"username" db:"username"`
	Deposit          float64 `json:"deposit" db:"deposit"`
	Withdraw         float64 `json:"withdraw" db:"withdraw"`
	Dividend         float64 `json:"dividend" db:"dividend"`
	Rebate           float64 `json:"rebate" db:"rebate"`
	NetAmount        float64 `json:"net_amount" db:"net_amount"`
	TY               string  `json:"ty" db:"ty"`
	ZR               string  `json:"zr" db:"zr"`
	QP               string  `json:"qp" db:"qp"`
	DJ               string  `json:"dj" db:"dj"`
	DZ               string  `json:"dz" db:"dz"`
	CP               string  `json:"cp" db:"cp"`
	FC               string  `json:"fc" db:"fc"`
	BY               string  `json:"by" db:"by"`
	CGHighRebate     string  `json:"cg_high_rebate" db:"cg_high_rebate"`         //CG高频彩返点
	CGOfficialRebate string  `json:"cg_official_rebate" db:"cg_official_rebate"` //CG官方彩返点
}

type MemberAggData struct {
	MemCount       int `db:"mem_count" json:"mem_count"`
	RegistCountNew int `db:"regist_count" json:"regist_count"`
	ActiveCount    int `db:"active_count" json:"active_count"`
}

type MemberLoginLog struct {
	Username string `msg:"username" json:"username"`
	IPS      string `msg:"ips" json:"ips"`
	Device   string `msg:"device" json:"device"`
	DeviceNo string `msg:"device_no" json:"device_no"`
	Date     uint32 `msg:"date" json:"date"`
	Serial   string `msg:"serial" json:"serial"`
	Parents  string `msg:"parents" json:"parents"`
	IsRisk   int    `msg:"-" json:"is_risk"`
	Prefix   string `msg:"prefix" json:"prefix"`
}

type MessageTD struct {
	Ts       string `json:"ts" db:"ts"`               //会员站内信id
	MsgID    string `json:"message_id" db:"msg_id"`   //站内信id
	Username string `json:"username" db:"username"`   //会员名
	Title    string `json:"title" db:"title"`         //标题
	SubTitle string `json:"sub_title" db:"sub_title"` //标题
	Content  string `json:"content" db:"content"`     //内容
	IsTop    int    `json:"is_top" db:"is_top"`       //0不置顶 1置顶
	IsVip    int    `json:"is_vip" db:"is_vip"`       //0非vip站内信 1vip站内信
	Ty       int    `json:"ty" db:"ty"`               //1站内消息 2活动消息
	IsRead   int    `json:"is_read" db:"is_read"`     //是否已读 0未读 1已读
	SendName string `json:"send_name" db:"send_name"` //发送人名
	SendAt   int64  `json:"send_at" db:"send_at"`     //发送时间
	Prefix   string `json:"prefix" db:"prefix"`       //商户前缀
}

type MessageTDData struct {
	T int64       `json:"t"`
	S int         `json:"s"`
	D []MessageTD `json:"d"`
}

type BalanceTransaction struct {
	Uid          string  `json:"uid"`
	Amount       float64 `json:"amount"`
	BeforeAmount float64 `json:"before_amount"`
	CashType     int     `json:"cash_type"`
	CreatedAt    int64   `json:"created_at"`
	BillNo       string  `json:"bill_no"`
	AfterAmount  float64 `json:"after_amount"`
	Username     string  `json:"username"`
}

type WaterFlow struct {
	UID                 string `json:"uid" redis:"uid"`
	Username            string `json:"username" redis:"username"`
	IsDowngrade         int    `json:"is_downgrade" redis:"is_downgrade"`                   //是否降级
	TotalDeposit        string `json:"total_deposit" redis:"total_deposit"`                 //累计存款
	TotalWaterFlow      string `json:"total_water_flow" redis:"total_water_flow"`           //累计流水
	ReturnDeposit       string `json:"return_deposit" redis:"return_deposit"`               //回归流水
	ReturnWaterFlow     string `json:"return_water_flow" redis:"return_water_flow"`         //回归存款
	RelegationWaterFlow string `json:"relegation_water_flow" redis:"relegation_water_flow"` //保级流水
}

/*
 * @Description: MemberCardList 会员银行卡 数据概览表结构
 * @Author: starc
 * @Date: 2022/6/1 12:38
 * @LastEditTime: 2022/6/1 20:00
 * @LastEditors: starc
 */
type MemberCardOverviewData struct {
	Username string `rule:"none" name:"username" msg:"username error"`
	BankName string `rule:"none" name:"bankname" msg:"bankname error"`
	BankNo   string `rule:"none" name:"bankno" msg:"bankno error"`
	RealName string `rule:"none" name:"realname" msg:"realname error"`
	Ip       string `rule:"none" name:"ip" msg:"ip error"`
	Status   int    `rule:"digit" min:"0" max:"1" default:"1" msg:"status error"`
	Ts       string `rule:"none" nane:"ts" `
}
