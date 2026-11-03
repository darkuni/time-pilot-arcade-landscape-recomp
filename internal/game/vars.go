package game

// Named work-RAM variables.  Addresses are those of the original program so
// that state can be compared byte-for-byte with the reference machine.
const (
	vObj0       = 0xa800 // player record (16 bytes); [0] = state, [2] = heading, ...
	vFrameCtr   = 0xa980 // free-running frame counter
	vCoin1Cnt   = 0xa981 // credits-in-progress counters
	vCoin2Cnt   = 0xa982
	vCoinSh2    = 0xa983 // button edge history shift registers
	vCoin1Tmr   = 0xa984 // coin counter pulse timers
	vCoin2Tmr   = 0xa985
	vCredits    = 0xa986 // BCD credit count
	vFlip       = 0xa987 // 1 = normal orientation
	vHiScore    = 0xa98b // 3 bytes BCD, little endian (a98b..a98d)
	vInCoin     = 0xa9ad // inverted input latches (see nmi)
	vInStart    = 0xa9ae
	vInP1       = 0xa9af
	vInP2       = 0xa9b0
	vInDsw0     = 0xa9b1
	vQWr        = 0xa9b2 // draw-queue write index
	vQRd        = 0xa9b3 // draw-queue read index
	vState      = 0xa9ab // top-level game state (0 boot,1 attract,2 start,3 play)
	vSub        = 0xa9ac // sub-state
	vSubTimer   = 0xa9eb
	vNmiTick    = 0xa9ce
	vSndQ       = 0xac43 // sound command queue: [0]=count, [1..]=bytes
	vQueue      = 0xac00 // draw command ring buffer (32 x 2 bytes)
	vGameActive = 0xad30 // nonzero while a game is in progress
	vPlayer2    = 0xad32 // 0 = player 1 active, 1 = player 2 active
	vScoreP1    = 0xad33 // 3 bytes BCD little endian
	vScoreP2    = 0xad36
	vStage      = 0xad04
	vShadowSpr  = 0xaa00 // shadow sprite tables
)
