package hunt

// State represents a Hunt FSM state.
// Each state defines which tools are available and the state-specific data.
type State interface {
	// Name returns the state name (INITIALIZE, TRACE, SCAN, EXPOSE, LOCK).
	Name() string
	// AllowedTools returns the list of tool names available in this state.
	AllowedTools() []string
}

// InitializeState represents the INITIALIZE phase.
// Purpose: Register evidence source and validate case workspace.
type InitializeState struct {
	CaseID       string // UUID v4
	SandboxRoot  string // /cases/<case_id>/
	EvidenceKind string // disk_image|memory_dump|pcap|log_directory
	EvidencePath string // /evidence/...
}

func (s *InitializeState) Name() string {
	return "INITIALIZE"
}

func (s *InitializeState) AllowedTools() []string {
	return []string{"case.open"}
}

// TraceState represents the TRACE phase.
// Purpose: Build forensic timeline and identify candidate events.
type TraceState struct {
	CaseID       string        // from INITIALIZE
	TimelinePath string        // /cases/<case_id>/timeline.jsonl
	EventCount   int           // de-duplicated events
	SourcesFound []string      // disk|memory|logs|network
	LastQuery    string        // human-readable last query
	QueryResults []interface{} // recent log hits
}

func (s *TraceState) Name() string {
	return "TRACE"
}

func (s *TraceState) AllowedTools() []string {
	return []string{"timeline.build", "log.query"}
}

// ScanState represents the SCAN phase.
// Purpose: Run indicator-of-compromise detection and memory analysis.
type ScanState struct {
	CaseID      string
	ScanID      string        // UUID v4, per IOC scan
	Matches     []Finding     // IOC matches
	Processes   []Process     // from memory.process_list
	MalfindsVAD []VADRegion   // from memory.malfind
	NetFlows    []NetworkFlow // from net.flow_summary
	ToolsRun    []string      // which tools succeeded
}

func (s *ScanState) Name() string {
	return "SCAN"
}

func (s *ScanState) AllowedTools() []string {
	return []string{
		"iocs.scan",
		"memory.process_list",
		"memory.malfind",
		"net.flow_summary",
	}
}

// Finding represents an IOC match.
type Finding struct {
	Rule     string // YARA/Sigma rule name
	Path     string // file path or memory address
	Offset   int    // offset in file/memory
	Severity string // info|low|medium|high|critical
	Context  string // contextual information
}

// Process represents a process entry from memory analysis.
type Process struct {
	PID      int
	Name     string
	Path     string
	Cmdline  string
	Timestamp string
}

// VADRegion represents a Virtual Address Descriptor region (memory injection).
type VADRegion struct {
	Address    string // hex address
	Size       int    // bytes
	Protection string // RWX flags
	Type       string // Image|Mapped|Private
}

// NetworkFlow represents a network flow summary.
type NetworkFlow struct {
	SrcIP    string
	DstIP    string
	SrcPort  int
	DstPort  int
	Protocol string // TCP|UDP
	Packets  int
	Bytes    int
}

// ExposeState represents the EXPOSE phase.
// Purpose: Cross-verify findings and build confidence scores.
type ExposeState struct {
	CaseID      string
	Findings    []VerifiedFinding // with confidence, verdict
	ReportPath  string             // /cases/<case_id>/report/findings.jsonl
	FindingHash string             // SHA256 of report for chain-of-custody
}

func (s *ExposeState) Name() string {
	return "EXPOSE"
}

func (s *ExposeState) AllowedTools() []string {
	return []string{"verify.cross_check", "report.append"}
}

// VerifiedFinding represents a verified finding with confidence score.
type VerifiedFinding struct {
	FindingID   string   // UUID
	Title       string
	Severity    string  // info|low|medium|high|critical
	Narrative   string
	Confidence  float64 // 0.0 - 1.0
	Verdict     string  // confirmed|unconfirmed|false_positive
	Evidence    []string
	MITREAttack []string
	VerifiedAt  string // ISO 8601
}

// LockState represents the LOCK phase (final state).
// Purpose: Finalize case, generate report, clean up investigation state.
type LockState struct {
	CaseID    string // from prior states
	Status    string // CLOSED|ERROR|INCOMPLETE
	ReportPath string
	Summary    string
	ClosedAt  string // ISO 8601
}

func (s *LockState) Name() string {
	return "LOCK"
}

func (s *LockState) AllowedTools() []string {
	return []string{} // No tools in LOCK phase
}
