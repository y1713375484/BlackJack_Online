package game

import (
	"fmt"
	"math/rand"
	"time"
)

type Game struct{}

var Poker = map[string]int{
	"♠A": 0, "♠2": 2, "♠3": 3, "♠4": 4, "♠5": 5, "♠6": 6, "♠7": 7, "♠8": 8, "♠9": 9, "♠10": 10, "♠J": 10, "♠Q": 10, "♠K": 10,
	"♣A": 0, "♣2": 2, "♣3": 3, "♣4": 4, "♣5": 5, "♣6": 6, "♣7": 7, "♣8": 8, "♣9": 9, "♣10": 10, "♣J": 10, "♣Q": 10, "♣K": 10,
	"♥A": 0, "♥2": 2, "♥3": 3, "♥4": 4, "♥5": 5, "♥6": 6, "♥7": 7, "♥8": 8, "♥9": 9, "♥10": 10, "♥J": 10, "♥Q": 10, "♥K": 10,
	"♦A": 0, "♦2": 2, "♦3": 3, "♦4": 4, "♦5": 5, "♦6": 6, "♦7": 7, "♦8": 8, "♦9": 9, "♦10": 10, "♦J": 10, "♦Q": 10, "♦K": 10,
}

// 初始化牌局
func (this *Game) Init(GameUserPoker map[int][]map[string]int, OtherPoker map[string]int, OtherPokerKeys *[]string) {
	keys := []string{}
	for k := range Poker {
		keys = append(keys, k)
	}
	shuffle(keys)
	//将洗过的牌一一对应到剩余牌堆里
	for l := 0; l < len(keys); l++ {
		OtherPoker[keys[l]] = Poker[keys[l]]
	}

	s := 0
	for k2 := range GameUserPoker {
		for i := 0; i < 2; i++ {
			GameUserPoker[k2] = append(GameUserPoker[k2], map[string]int{
				keys[s]: Poker[keys[s]],
			})
			//发过的牌从剩余牌中删除
			delete(OtherPoker, keys[s])
			s++
		}

	}

	*OtherPokerKeys = keys[s:]

}

// 给用户发牌
func (this *Game) SendPoker(userId int, GameUserPoker map[int][]map[string]int, OtherPoker map[string]int, OtherPokerKeys *[]string) (map[string]int, bool) {
	firstKey := (*OtherPokerKeys)[0]

	//新的牌
	newUserPoker := []map[string]int{{
		firstKey: OtherPoker[firstKey],
	}}
	//将旧的牌追加到新牌后面
	GameUserPoker[userId] = append(newUserPoker, GameUserPoker[userId]...)
	//fmt.Println("加牌后用户手牌", GameUserPoker[userId])
	//当前玩家所展示出的牌，也就是牌中出去最后一张牌的剩余牌
	facePoker := GameUserPoker[userId][:len(GameUserPoker[userId])-1]
	sum := 0
	for _, v := range facePoker {
		for _, v2 := range v {
			//碰到A的时候判断当前明牌牌面是否大于等于10，如果大于等于10就当1，
			//为什么要判断是否等于10呢？因为要牌后是3张牌，如果牌面已经是10点，
			//那么A再当11点加上没明的那张牌还是会爆
			if v2 == 0 {
				if sum >= 10 {
					sum += 1
				}
			} else {
				sum += v2
			}
		}
	}

	fmt.Println(sum)

	var ok bool
	if sum > 21 {
		ok = false
	} else {
		ok = true
	}

	newPoker := map[string]int{
		firstKey: OtherPoker[firstKey],
	}
	//从剩余牌索引中删除已经发了的牌
	*OtherPokerKeys = (*OtherPokerKeys)[1:]
	//从剩余牌中按照索引删除已经发了的牌
	delete(OtherPoker, firstKey)
	return newPoker, ok
}

/*
 * @Description: 洗牌方法
 * @param slice	要进行洗牌的切片
 */
func shuffle(slice []string) {
	rand.Seed(time.Now().UnixNano())

	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// 游戏最终比玩家牌点数
func (this *Game) GameFinal(GameUserPoker map[int][]map[string]int) string {
	//用户1总点数、用户1 A的数量，用户1牌数量
	user1Num, user1ANum, user1PokerNum, user2Num, user2ANum, user2PokerNum := 0, 0, 0, 0, 0, 0
	msg := ""
	//先计算除了A意外的牌值，并统计A的张数
	for k, v := range GameUserPoker {
		for _, v2 := range v {
			for _, v3 := range v2 {
				switch k {
				case 1:
					user1Num += v3
					if v3 == 0 {
						user1ANum += 1
					}
					user1PokerNum++
					break
				case 2:
					user2Num += v3
					if v3 == 0 {
						user2ANum += 1
					}
					user2PokerNum++
					break
				}
			}
		}
	}
	//fmt.Println(user1Num, user2Num, user1PokerNum, user2PokerNum)

	sumFunc := func(num int, ANum int) int {
		if ANum == 0 {
			return num
		}

		if ANum > 1 {
			if num > 9 {
				num += 1 * ANum
			} else {
				num += 11
				num += (ANum - 1) * 1
			}
		} else {
			if num <= 10 {
				num += 11
			} else {
				num += 1
			}
		}
		return num
	}

	user1Num = sumFunc(user1Num, user1ANum)
	user2Num = sumFunc(user2Num, user2ANum)
	//if user1Num > 9 {
	//	user1Num += 1 * user1ANum
	//} else {
	//	user1Num += 11
	//	user1Num += (user1ANum - 1) * 1
	//}
	//
	//if user2Num > 9 {
	//	user2Num += 1 * user2ANum
	//} else {
	//	user2Num += 11
	//	user2ANum += (user2ANum - 1) * user1ANum
	//}

	fmt.Println(user1Num, user2Num, user1PokerNum, user2PokerNum)

	if user2Num > 21 && user1Num > 21 {
		msg = fmt.Sprintf("平局，庄家为%v点，闲家为%v点", user1Num, user2Num)
		return msg
	}

	if user1Num <= 21 && user1PokerNum == 5 {
		msg = "庄家赢，庄家为五小"
		return msg
	}

	if user2Num <= 21 && user2PokerNum == 5 {
		msg = "闲家赢，闲家为五小"
		return msg
	}

	if user1Num > 21 {
		msg = fmt.Sprintf("闲家赢，庄家为%v点，闲家为%v点", user1Num, user2Num)
		return msg
	}

	if user2Num > 21 {
		msg = fmt.Sprintf("庄家赢，庄家为%v点，闲家为%v点", user1Num, user2Num)
		return msg
	}

	if user1Num-user2Num >= 0 {
		msg = fmt.Sprintf("庄家赢，庄家为%v点，闲家为%v点", user1Num, user2Num)
		return msg
	} else {
		msg = fmt.Sprintf("闲家赢，庄家为%v点，闲家为%v点", user1Num, user2Num)
		return msg
	}

}
