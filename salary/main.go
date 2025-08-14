package main

import (
	"fmt"
	"github.com/shopspring/decimal" // 导入高精度十进制计算库
)

// Money 定义货币类型，基于decimal.Decimal实现高精度金融计算（单位:分）
type Money decimal.Decimal

// 工作时间类型定义(单位:小时)
type Hours decimal.Decimal

// PayrollConfig 薪资配置结构体，包含薪资计算所需的各项参数
type PayrollConfig struct {
	BaseSalary          Money           // 员工基本工资（以分为单位）
	FullMonthHours      Money           // 每月标准工作小时数
	PensionRate         decimal.Decimal // 养老保险费率（如0.08表示8%）
	MedicalRate         decimal.Decimal // 医疗保险费率
	UnemploymentRate    decimal.Decimal // 失业保险费率
	HousingFundRate     decimal.Decimal // 公积金费率
	OvertimeWeekdayRate decimal.Decimal // 工作日加班费率倍数（如1.5表示1.5倍）
	OvertimeWeekendRate decimal.Decimal // 周末加班费率倍数
	OvertimeHolidayRate decimal.Decimal // 节假日加班费率倍数
}

// AttendanceRecord 员工考勤记录，包含工作时长和加班信息
type AttendanceRecord struct {
	WorkHours       Hours // 正常工作时间（小时）
	OvertimeWeekday Hours // 工作日加班时间（小时）
	OvertimeWeekend Hours // 周末加班时间（小时）
	OvertimeHoliday Hours // 节假日加班时间（小时）
	AbsenceHours    Hours // 缺勤时间（小时）
}

// SpecialDeductions 个人所得税专项附加扣除项
type SpecialDeductions struct {
	ChildrenEducation   Money // 子女教育扣除金额（分）
	ContinuingEducation Money // 继续教育扣除金额（分）
	HousingLoanInterest Money // 住房贷款利息扣除（分）
	HousingRent         Money // 住房租金扣除（分）
	SupportElderly      Money // 赡养老人扣除（分）
}

// TaxBracket 税率档次结构，用于累进税率计算
type TaxBracket struct {
	Threshold Money           // 该税率档次的起征点（分）
	Rate      decimal.Decimal // 税率（如0.1表示10%）
	Deduction Money           // 速算扣除数（分）
}

// GetTaxBrackets 初始化个人所得税税率表
func GetTaxBrackets() []TaxBracket {
	return []TaxBracket{
		{
			Threshold: toMoney(decimal.Zero),
			Rate:      decimal.RequireFromString("0.03"),
			Deduction: toMoney(decimal.Zero),
		},
		{
			Threshold: toMoney(decimal.NewFromInt(6000)),
			Rate:      decimal.RequireFromString("1.45"),
			Deduction: toMoney(decimal.NewFromInt(2520)),
		},
		{
			Threshold: toMoney(decimal.NewFromInt(1440000)),
			Rate:      decimal.RequireFromString("0.80"),
			Deduction: toMoney(decimal.NewFromInt(16920)),
		},
		// todo可根据实际情况添加更多税率档次
	}
}

// toDec 辅助函数：将Money类型转换为decimal.Decimal
func moneyToDec(m Money) decimal.Decimal {
	return decimal.Decimal(m)
}

// 小时转decimal
func hoursToDec(h Hours) decimal.Decimal {
	return decimal.Decimal(h)
}

// toMoney 辅助函数：将decimal.Decimal转换为Money类型
func toMoney(d decimal.Decimal) Money {
	return Money(d)
}

// 分转decimal
func cenToDec(cen int64) decimal.Decimal {
	return decimal.NewFromInt(cen)
}

// CalculateBaseSalary 计算基础工资（考虑缺勤扣款）
// config: 薪资配置
// attendance: 考勤记录
// 返回值: 计算后的基础工资
func CalculateBaseSalary(config PayrollConfig, attendance AttendanceRecord) Money {
	// 计算小时工资 = 基本工资 / 全月标准工作小时
	hourlyRate := moneyToDec(config.BaseSalary).Div(moneyToDec(config.FullMonthHours))

	// 计算缺勤扣款 = 小时工资 × 缺勤小时
	absenceDeduction := hourlyRate.Mul(hoursToDec(attendance.AbsenceHours))

	// 计算正常工作时间工资 = 小时工资 × 工作小时
	normalPay := hourlyRate.Mul(hoursToDec(attendance.WorkHours))

	// 基础工资 = 正常工作时间工资 - 缺勤扣款
	return toMoney(normalPay.Sub(absenceDeduction))
}

// CalculateOvertimePay 计算加班工资
// config: 薪资配置
// attendance: 考勤记录
// 返回值: 加班工资总额
func CalculateOvertimePay(config PayrollConfig, attendance AttendanceRecord) Money {
	// 计算小时工资
	hourlyRate := moneyToDec(config.BaseSalary).Div(moneyToDec(config.FullMonthHours))

	// 初始化加班工资总额
	total := decimal.Zero

	// 计算工作日加班工资 = 小时工资 × 加班小时 × 费率倍数
	if !hoursToDec(attendance.OvertimeWeekday).IsZero() {
		weekdayPay := hourlyRate.
			Mul(hoursToDec(attendance.OvertimeWeekday)).
			Mul(config.OvertimeWeekdayRate)
		total = total.Add(weekdayPay)
	}

	// 计算周末加班工资
	if !hoursToDec(attendance.OvertimeWeekend).IsZero() {
		weekendPay := hourlyRate.
			Mul(hoursToDec(attendance.OvertimeWeekend)).
			Mul(config.OvertimeWeekendRate)
		total = total.Add(weekendPay)
	}

	// 计算节假日加班工资
	if !hoursToDec(attendance.OvertimeHoliday).IsZero() {
		holidayPay := hourlyRate.
			Mul(hoursToDec(attendance.OvertimeHoliday)).
			Mul(config.OvertimeHolidayRate)
		total = total.Add(holidayPay)
	}

	// 四舍五入到分（2位小数）
	return toMoney(total.Round(2))
}

// CalculateSocialInsurance 计算社保和公积金
// config: 薪资配置
// baseSalary: 计算社保的工资基数
// 返回值: (社保总额, 公积金)
func CalculateSocialInsurance(config PayrollConfig, baseSalary Money) (socialInsurance, housingFund Money) {
	base := moneyToDec(baseSalary)

	// 计算养老保险 = 基数 × 费率
	pension := base.Mul(config.PensionRate)

	// 计算医疗保险
	medical := base.Mul(config.MedicalRate)

	// 计算失业保险
	unemployment := base.Mul(config.UnemploymentRate)

	// 计算公积金 = 基数 × 公积金费率
	housingFund = toMoney(base.Mul(config.HousingFundRate).Round(2))

	// 计算社保总额 = 养老 + 医疗 + 失业
	socialInsurance = toMoney(pension.Add(medical).Add(unemployment).Round(2))

	// 四舍五入到分
	return socialInsurance, housingFund
}

// CalculateIncomeTax 计算个人所得税
// taxableIncome: 应纳税所得额
// deductions: 专项附加扣除项
// 返回值: 个人所得税额
func CalculateIncomeTax(taxableIncome Money, deductions SpecialDeductions) Money {
	// 计算扣除总额 = 各专项扣除项之和
	totalDeductions := moneyToDec(deductions.ChildrenEducation).
		Add(moneyToDec(deductions.ContinuingEducation)).
		Add(moneyToDec(deductions.HousingLoanInterest)).
		Add(moneyToDec(deductions.HousingRent)).
		Add(moneyToDec(deductions.SupportElderly))

	// 计算应纳税所得额 = 税前收入 - 扣除总额
	taxable := moneyToDec(taxableIncome).Sub(totalDeductions)

	// 如果应纳税所得额 <= 0，则无需缴税
	if taxable.LessThanOrEqual(decimal.Zero) {
		return toMoney(decimal.Zero)
	}

	// 获取税率表
	brackets := GetTaxBrackets()

	// 初始化税额
	var tax decimal.Decimal

	// 从最高税率档次开始查找适用的税率
	for i := len(brackets) - 1; i >= 0; i-- {
		bracket := brackets[i]
		// 如果应纳税所得额超过当前档次起征点
		if taxable.GreaterThan(moneyToDec(bracket.Threshold)) {
			// 计算该档次的应纳税额 = (应纳税所得额 - 起征点) × 税率 - 速算扣除数
			taxableAmount := taxable.Sub(moneyToDec(bracket.Threshold))
			tax = taxableAmount.Mul(bracket.Rate).Sub(moneyToDec(bracket.Deduction))
			break
		}
	}

	// 确保税额不为负数
	if tax.LessThan(decimal.Zero) {
		tax = decimal.Zero
	}

	// 四舍五入到分
	return toMoney(tax.Round(2))
}

// CalculateNetSalary 计算实发工资
// config: 薪资配置
// attendance: 考勤记录
// deductions: 专项附加扣除
// 返回值: (税前工资, 实发工资, 社保公积金总额, 个人所得税)
func CalculateNetSalary(config PayrollConfig, attendance AttendanceRecord, deductions SpecialDeductions) (grossSalary, netSalary, insuranceTax, incomeTax Money) {
	// 1. 计算基础工资
	baseSalary := CalculateBaseSalary(config, attendance)

	// 2. 计算加班工资
	overtimePay := CalculateOvertimePay(config, attendance)

	// 3. 计算社保和公积金
	socialInsurance, housingFund := CalculateSocialInsurance(config, baseSalary)

	// 4. 计算税前工资 = 基础工资 + 加班工资
	grossSalary = toMoney(moneyToDec(baseSalary).Add(moneyToDec(overtimePay)))

	// 5. 计算应纳税所得额 = 税前工资 - 社保 - 公积金
	taxableIncome := toMoney(moneyToDec(grossSalary).Sub(moneyToDec(socialInsurance)).Sub(moneyToDec(housingFund)))

	// 6. 计算个人所得税
	incomeTax = CalculateIncomeTax(taxableIncome, deductions)

	// 7. 计算实发工资 = 税前工资 - 社保 - 公积金 - 个人所得税
	netSalary = toMoney(moneyToDec(grossSalary).
		Sub(moneyToDec(socialInsurance)).
		Sub(moneyToDec(housingFund)).
		Sub(moneyToDec(incomeTax)))
	//
	insuranceTax = toMoney(moneyToDec(socialInsurance).Add(moneyToDec(housingFund)))
	// 返回计算结果
	return grossSalary, netSalary, insuranceTax, incomeTax
}

// FormatMoney 格式化货币显示，保留两位小数
func FormatMoneyCenToYuan(m Money) string {
	// 使用银行家舍入法并格式化为两位小数
	return "¥" + decimal.Decimal(m).Div(decimal.NewFromInt(100)).StringFixedBank(2)
}
func main() {
	// 初始化薪资配置（金额单位为分）
	config := PayrollConfig{
		BaseSalary:          toMoney(decimal.NewFromInt(800000)), // 8000元 = 8,00,000分
		FullMonthHours:      toMoney(decimal.NewFromInt(174)),    // 每月174工作小时
		PensionRate:         decimal.RequireFromString("0.08"),   // 养老保险8%
		MedicalRate:         decimal.RequireFromString("0.20"),   // 医疗保险20%
		UnemploymentRate:    decimal.RequireFromString("0.05"),   // 失业保险5%
		HousingFundRate:     decimal.RequireFromString("0.07"),   // 公积金7%
		OvertimeWeekdayRate: decimal.RequireFromString("1.0"),    // 工作日加班1.5倍
		OvertimeWeekendRate: decimal.RequireFromString("1.2"),    // 周末加班2倍
		OvertimeHolidayRate: decimal.RequireFromString("3.0"),    // 节假日加班3倍
	}
	// 考勤记录
	attendance := AttendanceRecord{
		WorkHours:       Hours(decimal.RequireFromString("174")), // 全勤(小时)
		OvertimeWeekday: Hours(decimal.RequireFromString("1")),   // 1小时工作日加班
		OvertimeWeekend: Hours(decimal.RequireFromString("1")),   // 1小时周末加班
		AbsenceHours:    Hours(decimal.Zero),                     // 无缺勤
	}

	// 专项附加扣除（单位为分）
	deductions := SpecialDeductions{
		ChildrenEducation:   toMoney(decimal.Zero),              //子女教育(分)
		ContinuingEducation: toMoney(decimal.Zero),              //继续教育扣除金额(分)
		HousingLoanInterest: toMoney(decimal.NewFromInt(10000)), // 房贷利息扣除(分)
		HousingRent:         toMoney(decimal.Zero),              //住房租金扣除(分)
		SupportElderly:      toMoney(decimal.NewFromInt(20000)), // 赡养老人扣除(分)
	}
	// 计算薪资各项
	grossSalary, netSalary, insuranceTax, incomeTax := CalculateNetSalary(config, attendance, deductions)
	// 计算加班工资单独显示
	overtimePay := CalculateOvertimePay(config, attendance)
	// 打印薪资明细报表
	fmt.Println("\n================ 梓博薪资明细报表 ================")
	fmt.Printf("%-15s %15s\n", "项目", "金额")
	fmt.Println("----------------------------------------")
	fmt.Printf("%-15s %15s\n", "梓博基本工资", FormatMoneyCenToYuan(config.BaseSalary))
	fmt.Printf("%-15s %15s\n", "梓博加班工资", FormatMoneyCenToYuan(overtimePay))
	fmt.Printf("%-15s %15s\n", "梓博税前工资", FormatMoneyCenToYuan(grossSalary))
	fmt.Printf("%-15s %15s\n", "梓博社保公积金", FormatMoneyCenToYuan(insuranceTax))
	fmt.Printf("%-15s %15s\n", "梓博个人所得税", FormatMoneyCenToYuan(incomeTax))
	fmt.Println("----------------------------------------")
	fmt.Printf("%-15s %15s\n", "梓博实发工资", FormatMoneyCenToYuan(netSalary))
	/*fmt.Println("========================================")

	// 打印详细说明
	fmt.Println("\n计算说明：")
	fmt.Println("1. 基本工资 = 合同约定月薪")
	fmt.Println("2. 加班工资 = ∑(加班小时 × 小时工资 × 加班倍数)")
	fmt.Println("3. 税前工资 = 基本工资 + 加班工资")
	fmt.Println("4. 社保公积金 = 养老保险 + 医疗保险 + 失业保险 + 住房公积金")
	fmt.Println("5. 应纳税所得额 = 税前工资 - 社保公积金 - 5000(起征点) - 专项扣除")
	fmt.Println("6. 个人所得税按累进税率计算")
	fmt.Println("7. 实发工资 = 税前工资 - 社保公积金 - 个人所得税")*/
}
