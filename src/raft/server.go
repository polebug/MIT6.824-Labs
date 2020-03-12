package raft

import (
	"labrpc"
	// "log"
)

// GetState return currentTerm and whether this server believes it is the leader.
func (rf *Raft) GetState() (term int, isleader bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	term = rf.CurrentTerm
	isleader = (rf.state == LEADER)
	// log.Printf("[GetState] node = %v, term = %v, state = %v \n", rf.me, term, rf.state)
	return term, isleader
}

// Start lab 3 k/v server’s RPC handlers(e.g. Put/Get) call Start().
// start Raft agreement on a new log entry. Start() returns immediately —
// RPC handler must then wait for commit.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
// even if the Raft instance has been killed, this function should return gracefully.
func (rf *Raft) Start(command interface{}) (index int, term int, isLeader bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if isLeader = (rf.state == LEADER); isLeader {
		term = rf.CurrentTerm
		index = len(rf.Logs) // initialized to 1
		rf.Logs = append(rf.Logs, LogEntry{
			Command: command,
			Term:    term,
		})
		// log.Printf("[Start]: log = %v \n", rf.Logs)
		rf.matchIndex[rf.me] = len(rf.Logs) - 1
		rf.nextIndex[rf.me] = len(rf.Logs)
		rf.persist()
	}

	return index, term, isLeader
}

// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

// Make create a Raft server
func Make(
	peers []*labrpc.ClientEnd, // the ports of all the Raft servers, all the servers' peers[] arrays have the same order
	me int, // this server's port is peers[me]
	persister *Persister, // is a place for this server to save its persistent state, and also initially holds the most recent saved state
	applyCh chan ApplyMsg, // a channel on which the tester or service expects Raft to send ApplyMsg messages
) *Raft {
	// initialization
	rf := &Raft{}
	rf.mu.Lock()
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	rf.SetFollower(0)
	rf.Logs = make([]LogEntry, 1)

	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.applyCh = applyCh

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	rf.matchIndex = make([]int, len(peers))
	rf.nextIndex = make([]int, len(peers))

	for i := 0; i < len(peers); i++ {
		rf.nextIndex[i] = len(rf.Logs)
	}
	rf.mu.Unlock()

	// create a background goroutine that will kick off leader election periodically
	go rf.ElectionTimeOut()
	// check committed logs in real time and apply commands to the state machined
	go rf.WaitCmdApplied()

	return rf
}
