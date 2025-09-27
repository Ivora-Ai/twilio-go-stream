package dectector

import (
	"fmt"
	"math/rand"
	"time"
)

var COOLING_PERIOD_INTRRUPT = 10 * time.Second

var COOLING_PERIOD_NO_ONE_SPOKE = 15 * time.Second
var NO_ONE_SPOKE_IN_LAST_X_SEC = 15 * time.Second

type Interrupt struct {
	AgentSpeaking              bool
	UserSpeaking               bool
	LastEventFiredAt           time.Time
	LastNoOneSpokeEventFiredAt time.Time
	LastTimeAgentSpoke         time.Time
	LastTimeUserSpoke          time.Time
	callStartedAt              time.Time
	CallDuration               int
	AgentResponse              func(bool, string)
}

func (i *Interrupt) AgentSpoke(b bool) { //attach to AgentResponse where audio is pused to ws
	i.AgentSpeaking = b

	i.LastTimeAgentSpoke = time.Now()

}
func (i *Interrupt) UserSpoke(b bool) { //attach to interim result
	i.UserSpeaking = b

	i.LastTimeUserSpoke = time.Now()

}

func (i *Interrupt) Reset() {
	i.AgentSpeaking = false
	i.UserSpeaking = false
	// i.LastEventFiredAt = time.Now()
}
func (i *Interrupt) IsInterrupt() bool {
	return i.AgentSpeaking && i.UserSpeaking && !i.IsCooling()
}

func (i *Interrupt) IsCooling() bool { //returns false is system is cool again
	// coolling should be based on when user completes speaking right
	// it should be not cool till AgentResponse is called for LLM response only, once it is called system should be cool
	return time.Since(i.LastEventFiredAt) < COOLING_PERIOD_INTRRUPT
}

func (i *Interrupt) NoOneSpokeCooling() bool {
	return time.Since(i.LastNoOneSpokeEventFiredAt) < COOLING_PERIOD_NO_ONE_SPOKE
}

// GenerateRandomNumber returns a random number between min and max (inclusive)
func GenerateRandomNumber(min, max int) int {
	rand.Seed(time.Now().UnixNano()) // Seed to ensure randomness
	return rand.Intn(max-min+1) + min
}
func (i *Interrupt) FireInterrupt() {
	if i.IsCooling() {
		return
	}
	go func() {

		// min, max := 0, 5 // Define range
		// fillerWords := []string{
		// 	// English Filler Words
		// 	"alright", "okay", "umm", "uh", "hmm", "you know", "like", "I mean", "well",
		// 	"actually", "so", "kind of", "sort of", "basically", "I guess", "let’s see",
		// 	"you see", "I suppose", "by the way", "right",

		// 	// Hindi Filler Words
		// 	"अच्छा", "हाँ", "मतलब", "जैसे", "वैसे", "उम", "हूँ", "तो", "देखो", "समझे",
		// 	"चलो", "बात ये है कि", "सही है", "वो क्या कहते हैं", "मेरा मतलब है", "थोड़ा सा",
		// 	"कुछ ऐसा",
		// }
		// max = len(fillerWords) - 1
		// randomNumber := GenerateRandomNumber(min, max)

		i.AgentResponse(false, "yes")
	}()

	fmt.Println("Interrupt Fired -----------------------------------------------------------")
	i.LastEventFiredAt = time.Now()
}

func (i *Interrupt) DidNoOneSpokeInLastXSec() bool {

	if time.Since(i.LastTimeAgentSpoke) > NO_ONE_SPOKE_IN_LAST_X_SEC && time.Since(i.LastTimeUserSpoke) > NO_ONE_SPOKE_IN_LAST_X_SEC {
		// i.AgentResponse(false, "How can I help you?")
		return !i.AgentSpeaking && !i.UserSpeaking
	}
	return false
}

func (i *Interrupt) FireNoOneSpokeInLastXSec() {
	if i.NoOneSpokeCooling() {
		return
	}
	go func() {
		time.Sleep(1 * time.Second)
		i.AgentResponse(false, "are you still there?")
	}()

	fmt.Println("No one spoke Fired -----------------------------------------------------------")
	i.LastNoOneSpokeEventFiredAt = time.Now()
}

func (i *Interrupt) InterruptsManager() {
	go func() {
		for {
			if i.IsInterrupt() {
				i.FireInterrupt()
			}
			if i.DidNoOneSpokeInLastXSec() {
				i.FireNoOneSpokeInLastXSec()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (i *Interrupt) Manager(stopChan chan struct{}) {
	//add gracefull close
	// stopChan := make(chan struct{})
	// InterruptManager(stopChan)
	// close(stopChan)

	i.callStartedAt = time.Now()
	i.CallDurationManager(stopChan)
	i.InterruptsManager()
	i.NoOneSpokeManager()

}

func (i *Interrupt) NoOneSpokeManager() {
	go func() {
		for {

			if i.DidNoOneSpokeInLastXSec() {
				i.FireNoOneSpokeInLastXSec()
			}
			fmt.Println("Duration: ", i.CallDuration)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (i *Interrupt) CallDurationManager(stopChan chan struct{}) {
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				i.CallDuration = int(time.Since(i.callStartedAt).Seconds())
				if i.CallDuration > 280 {
					i.AgentResponse(false, "Thank you for calling, Goodbye")
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}
