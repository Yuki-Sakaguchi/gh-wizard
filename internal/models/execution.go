package models

import (
	"time"
)

// TaskStatus は実行タスクの状態を表す
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusInProgress
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusSkipped
)

// String はタスクステータスの文字列表現を返す
func (ts TaskStatus) String() string {
	switch ts {
	case TaskStatusPending:
		return "待機中"
	case TaskStatusInProgress:
		return "実行中"
	case TaskStatusCompleted:
		return "完了"
	case TaskStatusFailed:
		return "失敗"
	case TaskStatusSkipped:
		return "スキップ"
	default:
		return "不明"
	}
}

// GetIcon はタスクステータスのアイコンを返す
func (ts TaskStatus) GetIcon() string {
	switch ts {
	case TaskStatusPending:
		return "⬜"
	case TaskStatusInProgress:
		return "🔄"
	case TaskStatusCompleted:
		return "✅"
	case TaskStatusFailed:
		return "❌"
	case TaskStatusSkipped:
		return "⏭️"
	default:
		return "❓"
	}
}

// ExecutionTask は実行する個別タスクを表す
type ExecutionTask struct {
	ID            string
	Name          string
	Description   string
	Status        TaskStatus
	Progress      float64 // 0.0 - 1.0
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Error         error
	Weight        float64 // 全体進捗への重み（合計1.0）
	EstimatedTime time.Duration
}

// IsCompleted はタスクが完了しているかを判定する
func (et *ExecutionTask) IsCompleted() bool {
	return et.Status == TaskStatusCompleted || et.Status == TaskStatusSkipped
}

// IsFailed はタスクが失敗しているかを判定する
func (et *ExecutionTask) IsFailed() bool {
	return et.Status == TaskStatusFailed
}

// GetElapsedTime はタスクの経過時間を取得する
func (et *ExecutionTask) GetElapsedTime() time.Duration {
	if et.StartTime.IsZero() {
		return 0
	}

	if et.IsCompleted() || et.IsFailed() {
		return et.Duration
	}

	return time.Since(et.StartTime)
}

// ExecutionPlan は実行計画全体を管理する
type ExecutionPlan struct {
	Tasks           []ExecutionTask
	CurrentTaskID   string
	OverallProgress float64
	StartTime       time.Time
	EstimatedTotal  time.Duration
	Status          string
	Error           error
}

// NewExecutionPlan は実行計画を作成する
func NewExecutionPlan(state *WizardState) *ExecutionPlan {
	tasks := []ExecutionTask{
		{
			ID:            "validate",
			Name:          "設定の検証",
			Description:   "入力された設定内容を検証しています",
			Status:        TaskStatusPending,
			Weight:        0.1,
			EstimatedTime: 2 * time.Second,
		},
		{
			ID:            "create_repo",
			Name:          "リポジトリ作成",
			Description:   "GitHubにリポジトリを作成しています",
			Status:        TaskStatusPending,
			Weight:        0.3,
			EstimatedTime: 5 * time.Second,
		},
	}

	// テンプレート使用時の追加タスク
	if state.UseTemplate && state.SelectedTemplate != nil {
		tasks = append(tasks, ExecutionTask{
			ID:            "setup_template",
			Name:          "テンプレート設定",
			Description:   "選択されたテンプレートを適用しています",
			Status:        TaskStatusPending,
			Weight:        0.2,
			EstimatedTime: 3 * time.Second,
		})
	}

	// README作成タスク
	if state.RepoConfig.AddReadme {
		tasks = append(tasks, ExecutionTask{
			ID:            "create_readme",
			Name:          "README作成",
			Description:   "README.mdファイルを生成しています",
			Status:        TaskStatusPending,
			Weight:        0.15,
			EstimatedTime: 2 * time.Second,
		})
	}

	// クローンタスク
	if state.RepoConfig.SholdClone {
		tasks = append(tasks, ExecutionTask{
			ID:            "clone_repo",
			Name:          "ローカルクローン",
			Description:   "リポジトリをローカルにクローンしています",
			Status:        TaskStatusPending,
			Weight:        0.25,
			EstimatedTime: 4 * time.Second,
		})
	}

	// 重みを正規化
	totalWeight := 0.0
	for _, task := range tasks {
		totalWeight += task.Weight
	}
	for i := range tasks {
		tasks[i].Weight = tasks[i].Weight / totalWeight
	}

	// 推定時間を計算
	estimatedTotal := time.Duration(0)
	for _, task := range tasks {
		estimatedTotal += task.EstimatedTime
	}

	return &ExecutionPlan{
		Tasks:          tasks,
		Status:         "初期化中",
		EstimatedTotal: estimatedTotal,
	}
}

// GetCurrentTask は現在実行中のタスクを取得する
func (ep *ExecutionPlan) GetCurrentTask() *ExecutionTask {
	if ep.CurrentTaskID == "" {
		return nil
	}

	for i := range ep.Tasks {
		if ep.Tasks[i].ID == ep.CurrentTaskID {
			return &ep.Tasks[i]
		}
	}

	return nil
}

// GetTaskByID はIDでタスクを取得する
func (ep *ExecutionPlan) GetTaskByID(id string) *ExecutionTask {
	for i := range ep.Tasks {
		if ep.Tasks[i].ID == id {
			return &ep.Tasks[i]
		}
	}
	return nil
}

// UpdateTaskStatus はタスクの状態を更新する
func (ep *ExecutionPlan) UpdateTaskStatus(taskID string, status TaskStatus, progress float64, err error) {
	task := ep.GetTaskByID(taskID)
	if task == nil {
		return
	}

	// 開始時刻の設定
	if status == TaskStatusInProgress && task.StartTime.IsZero() {
		task.StartTime = time.Now()
	}

	// 完了・失敗時の処理
	if (status == TaskStatusCompleted || status == TaskStatusFailed || status == TaskStatusSkipped) && !task.EndTime.IsZero() {
		task.EndTime = time.Now()
		if !task.StartTime.IsZero() {
			task.Duration = task.EndTime.Sub(task.StartTime)
		}
		progress = 1.0
	}

	task.Status = status
	task.Progress = progress
	task.Error = err

	// 全体の進捗を再計算
	ep.calculateOverallProgress()
}

// calculateOverallProgress は全体の進捗を計算する
func (ep *ExecutionPlan) calculateOverallProgress() {
	totalProgress := 0.0

	for _, task := range ep.Tasks {
		taskProgress := task.Progress
		if task.Status == TaskStatusCompleted || task.Status == TaskStatusSkipped {
			taskProgress = 1.0
		}
		totalProgress += taskProgress * task.Weight
	}

	ep.OverallProgress = totalProgress
}

// GetNextPendingTask は次の待機中タスクを取得する
func (ep *ExecutionPlan) GetNextPendingTask() *ExecutionTask {
	for i := range ep.Tasks {
		if ep.Tasks[i].Status == TaskStatusPending {
			return &ep.Tasks[i]
		}
	}
	return nil
}

// IsCompleted は全タスクが完了しているかを判定する
func (ep *ExecutionPlan) IsCompleted() bool {
	for _, task := range ep.Tasks {
		if task.Status != TaskStatusCompleted && task.Status != TaskStatusSkipped {
			return false
		}
	}
	return true
}

// HasFailed は失敗したタスクがあるかを判定する
func (ep *ExecutionPlan) HasFailed() bool {
	for _, task := range ep.Tasks {
		if task.Status == TaskStatusFailed {
			return true
		}
	}
	return false
}

// GetElapsedTime は全体の経過時間を取得する
func (ep *ExecutionPlan) GetElapsedTime() time.Duration {
	if ep.StartTime.IsZero() {
		return 0
	}
	return time.Since(ep.StartTime)
}

// GetEstimatedRemainingTime は推定残り時間を取得する
func (ep *ExecutionPlan) GetEstimatedRemainingTime() time.Duration {
	if ep.OverallProgress <= 0 {
		return ep.EstimatedTotal
	}

	elapsedTime := ep.GetElapsedTime()
	if elapsedTime <= 0 {
		return ep.EstimatedTotal
	}

	// 現在の進捗から残り時間を推定
	estimatedTotalTime := time.Duration(float64(elapsedTime) / ep.OverallProgress)
	remainingTime := estimatedTotalTime - elapsedTime

	if remainingTime < 0 {
		return 0
	}

	return remainingTime
}

// ExecutionMessage は実行中のメッセージ型
type ExecutionMessage struct {
	TaskID   string
	Status   TaskStatus
	Progress float64
	Error    error
	Message  string
}

// RepositoryCreationResult はリポジトリ作成の結果
type RepositoryCreationResult struct {
	Success       bool
	RepositoryURL string
	ClonePath     string
	Error         error
	Message       string
}
