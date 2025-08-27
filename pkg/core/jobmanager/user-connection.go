//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package jobmanager

//=============================================================================

type UserConnection struct {
	username       string
	connectionCode string
	connected      bool
	scheduledJob   *ScheduledJob
}

//=============================================================================

func (uc *UserConnection) key() string {
	return uc.username + ":" + uc.connectionCode
}

//=============================================================================

func (uc *UserConnection) allocateToJob(job *ScheduledJob) {
	uc.scheduledJob        = job
	job.job.UserConnection = uc.key()
}

//=============================================================================

func (uc *UserConnection) isAllocated() bool {
	return uc.scheduledJob != nil
}

//=============================================================================

func (uc *UserConnection) deallocate() {
	job := uc.scheduledJob
	job.job.UserConnection = ""
	uc.scheduledJob        = nil
}

//=============================================================================

func newUserConnection(username, connectionCode string, connected bool) *UserConnection {
	return &UserConnection{
		username      : username,
		connectionCode: connectionCode,
		connected     : connected,
	}
}

//=============================================================================
