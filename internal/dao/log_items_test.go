// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"fmt"
	"log/slog"
	"regexp"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestLogItemsFilter(t *testing.T) {
	uu := map[string]struct {
		q    string
		opts dao.LogOptions
		e    []int
		err  error
	}{
		"empty": {
			opts: dao.LogOptions{},
		},
		"pod-name": {
			q: "blee",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{0, 1, 2},
		},
		"container-name": {
			q: "c1",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{0, 1, 2},
		},
		"message": {
			q: "zorg",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{2},
		},
		"fuzzy": {
			q: "-f zorg",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{2},
		},
	}

	for k := range uu {
		u := uu[k]
		ii := dao.NewLogItems()
		ii.Add(
			dao.NewLogItem([]byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))),
			dao.NewLogItemFromString("Bumble bee tuna"),
			dao.NewLogItemFromString("Jean Batiste Emmanuel Zorg"),
		)
		t.Run(k, func(t *testing.T) {
			_, n := client.Namespaced(u.opts.Path)
			for _, i := range ii.Items() {
				i.Pod, i.Container = n, u.opts.Container
			}
			res, _, err := ii.Filter(0, u.q, false)
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, u.e, res)
			}
		})
	}
}

func TestLogItemsRender(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		e    string
	}{
		"empty": {
			opts: dao.LogOptions{},
			e:    "Testing 1,2,3...\n",
		},
		"container": {
			opts: dao.LogOptions{
				Container: "fred",
			},
			e: "[teal::b]fred[-::-] Testing 1,2,3...\n",
		},
		"pod-container": {
			opts: dao.LogOptions{
				Path:      "blee/fred",
				Container: "blee",
			},
			e: "[teal::]fred [teal::b]blee[-::-] Testing 1,2,3...\n",
		},
		"full": {
			opts: dao.LogOptions{
				Path:          "blee/fred",
				Container:     "blee",
				ShowTimestamp: true,
			},
			e: "[gray::b]2018-12-14T10:36:43.326972-07:00 [-::-][teal::]fred [teal::b]blee[-::-] Testing 1,2,3...\n",
		},
	}

	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	for k := range uu {
		ii := dao.NewLogItems()
		ii.Add(dao.NewLogItem(s))
		u := uu[k]
		_, n := client.Namespaced(u.opts.Path)
		ii.Items()[0].Pod, ii.Items()[0].Container = n, u.opts.Container
		t.Run(k, func(t *testing.T) {
			res := make([][]byte, 1)
			ii.Render(0, u.opts.ShowTimestamp, res)
			assert.Equal(t, u.e, string(res[0]))
		})
	}
}

// BenchmarkLogItemsFilter 测试正则表达式匹配的性能
func BenchmarkLogItemsFilter(b *testing.B) {
	// 创建大量日志项用于测试
	ii := dao.NewLogItems()
	logLines := []string{
		"2018-12-14T10:36:43.326972-07:00 INFO Starting application server on port 8080",
		"2018-12-14T10:36:44.123456-07:00 DEBUG Database connection established successfully",
		"2018-12-14T10:36:45.789012-07:00 WARN High memory usage detected: 85%",
		"2018-12-14T10:36:46.456789-07:00 ERROR Failed to process request: timeout occurred",
		"2018-12-14T10:36:47.234567-07:00 INFO User authentication successful for user@example.com",
		"2018-12-14T10:36:48.890123-07:00 DEBUG Cache hit ratio: 92.5%",
		"2018-12-14T10:36:49.567890-07:00 WARN Connection pool nearly exhausted: 95% utilization",
		"2018-12-14T10:36:50.345678-07:00 INFO Processing batch job with 10000 records",
		"2018-12-14T10:36:51.012345-07:00 ERROR Database query failed: connection timeout",
		"2018-12-14T10:36:52.678901-07:00 DEBUG Request processed in 125ms",
	}

	// 重复添加日志行以增加数据量
	for i := 0; i < 1000; i++ {
		for _, line := range logLines {
			item := dao.NewLogItemFromString(line)
			item.Pod = "test-pod"
			item.Container = "test-container"
			ii.Add(item)
		}
	}

	// 添加包含非ASCII字符的日志行
	nonAsciiLogLines := []string{
		"2018-12-14T10:36:53.123456-07:00 INFO 应用程序启动成功，端口：8080",
		"2018-12-14T10:36:54.234567-07:00 ERROR データベース接続エラーが発生しました",
		"2018-12-14T10:36:55.345678-07:00 WARN 메모리 사용량이 높습니다: 85%",
		"2018-12-14T10:36:56.456789-07:00 DEBUG Пользователь успешно авторизован",
		"2018-12-14T10:36:57.567890-07:00 INFO Traitement du lot avec 用户数据 处理完成",
		"2018-12-14T10:36:58.678901-07:00 ERROR 网络连接超时，请重试",
		"2018-12-14T10:36:59.789012-07:00 WARN システムリソースが不足しています",
		"2018-12-14T10:37:00.890123-07:00 DEBUG 성능 최적화가 완료되었습니다",
	}

	// 添加非ASCII字符的日志项
	for i := 0; i < 500; i++ {
		for _, line := range nonAsciiLogLines {
			item := dao.NewLogItemFromString(line)
			item.Pod = "test-pod"
			item.Container = "test-container"
			ii.Add(item)
		}
	}

	// 测试不同的查询模式
	queries := []string{
		"ERROR",                      // 简单匹配
		"timeout",                    // 部分匹配
		"[0-9]{4}-[0-9]{2}-[0-9]{2}", // 复杂正则表达式
		"INFO.*server",               // 复合模式
		"应用程序",                       // 中文匹配
		"データベース",                     // 日文匹配
		"메모리",                        // 韩文匹配
		"Пользователь",               // 俄文匹配
		"用户数据",                       // 中文关键词
		"システム.*不足",                   // 日文正则表达式
	}

	for _, query := range queries {
		b.Run(fmt.Sprintf("query_%s", query), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, err := ii.Filter(0, query, false)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestLogItemsFilterNonASCII 测试非ASCII字符的过滤功能
func TestLogItemsFilterNonASCII(t *testing.T) {
	ii := dao.NewLogItems()

	// 添加包含各种非ASCII字符的日志项
	testLogs := []string{
		"2018-12-14T10:36:43.326972-07:00 INFO 应用程序启动成功",
		"2018-12-14T10:36:44.326972-07:00 ERROR データベース接続エラー",
		"2018-12-14T10:36:45.326972-07:00 WARN 메모리 사용량 높음",
		"2018-12-14T10:36:46.326972-07:00 DEBUG Пользователь авторизован",
		"2018-12-14T10:36:47.326972-07:00 INFO Normal ASCII log message",
	}

	for _, log := range testLogs {
		item := dao.NewLogItemFromString(log)
		item.Pod = "test-pod"
		item.Container = "test-container"
		ii.Add(item)
	}

	testCases := []struct {
		name     string
		query    string
		expected []int
	}{
		{"中文匹配", "应用程序", []int{0}},
		{"日文匹配", "データベース", []int{1}},
		{"韩文匹配", "메모리", []int{2}},
		{"俄文匹配", "Пользователь", []int{3}},
		{"ASCII匹配", "Normal", []int{4}},
		{"错误级别匹配", "ERROR", []int{1}},
		{"多字符匹配", "启动成功", []int{0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := ii.Filter(0, tc.query, false)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result, "查询 '%s' 的结果不匹配", tc.query)
		})
	}
}

// BenchmarkLogItemsFilterComparison 对比优化前后的性能
func BenchmarkLogItemsFilterComparison(b *testing.B) {
	// 创建测试数据
	ii := dao.NewLogItems()
	logLines := []string{
		"2018-12-14T10:36:43.326972-07:00 INFO Starting application server on port 8080",
		"2018-12-14T10:36:44.123456-07:00 DEBUG Database connection established successfully",
		"2018-12-14T10:36:45.789012-07:00 WARN High memory usage detected: 85%",
		"2018-12-14T10:36:46.456789-07:00 ERROR Failed to process request: timeout occurred",
		"2018-12-14T10:36:47.234567-07:00 INFO User authentication successful for user@example.com",
		"2018-12-14T10:36:48.890123-07:00 DEBUG Cache hit ratio: 92.5%",
		"2018-12-14T10:36:49.567890-07:00 WARN Connection pool nearly exhausted: 95% utilization",
		"2018-12-14T10:36:50.345678-07:00 INFO Processing batch job with 10000 records",
		"2018-12-14T10:36:51.012345-07:00 ERROR Database query failed: connection timeout",
		"2018-12-14T10:36:52.678901-07:00 DEBUG Request processed in 125ms",
		"2018-12-14T10:36:53.123456-07:00 INFO 应用程序启动成功，端口：8080",
		"2018-12-14T10:36:54.234567-07:00 ERROR データベース接続エラーが発生しました",
		"2018-12-14T10:36:55.345678-07:00 WARN 메모리 사용량이 높습니다: 85%",
	}

	// 添加大量测试数据
	for i := 0; i < 1000; i++ {
		for _, line := range logLines {
			item := dao.NewLogItemFromString(line)
			item.Pod = "test-pod"
			item.Container = "test-container"
			ii.Add(item)
		}
	}

	// 测试查询
	testQueries := []string{
		"ERROR",
		"timeout",
		"应用程序",
		"データベース",
		"[0-9]{4}-[0-9]{2}-[0-9]{2}",
	}

	for _, query := range testQueries {
		b.Run(fmt.Sprintf("optimized_%s", query), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, err := ii.Filter(0, query, false)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// 模拟优化前的版本（使用字符串转换）
		b.Run(fmt.Sprintf("old_string_conversion_%s", query), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 模拟旧版本的字符串转换开销
				_, _, err := benchmarkOldStringConversion(ii, 0, query, false)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// benchmarkOldStringConversion 模拟优化前使用字符串转换的版本
func benchmarkOldStringConversion(l *dao.LogItems, index int, q string, showTime bool) (matches []int, indices [][]int, err error) {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil, nil, err
	}

	matches, indices = make([]int, 0, l.Len()), make([][]int, 0, l.Len())
	ll := make([][]byte, l.Len()-index)
	l.Lines(index, showTime, ll)

	for i, line := range ll {
		// 模拟旧版本：先转换为字符串再进行匹配
		lineStr := string(line)
		locs := rx.FindAllStringIndex(lineStr, -1)

		if locs == nil {
			continue
		}

		matches = append(matches, i)
		ii := make([]int, 0, 10)
		for _, loc := range locs {
			// 模拟旧版本的字符串索引到字节索引的转换开销
			startBytes := len([]byte(lineStr[:loc[0]]))
			endBytes := len([]byte(lineStr[:loc[1]]))
			for j := startBytes; j < endBytes; j++ {
				ii = append(ii, j)
			}
		}
		indices = append(indices, ii)
	}

	return matches, indices, nil
}
