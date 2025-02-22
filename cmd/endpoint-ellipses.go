// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/minio/internal/config"
	"github.com/minio/pkg/ellipses"
	"github.com/minio/pkg/env"
)

// This file implements and supports ellipses pattern for
// `minio server` command line arguments.

// Endpoint set represents parsed ellipses values, also provides
// methods to get the sets of endpoints.
type endpointSet struct {
	argPatterns []ellipses.ArgPattern
	endpoints   []string   // Endpoints saved from previous GetEndpoints().
	setIndexes  [][]uint64 // All the sets.
}

// Supported set sizes this is used to find the optimal
// single set size.
var setSizes = []uint64{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// getDivisibleSize - returns a greatest common divisor of
// all the ellipses sizes.
func getDivisibleSize(totalSizes []uint64) (result uint64) {
	gcd := func(x, y uint64) uint64 {
		for y != 0 {
			x, y = y, x%y
		}
		return x
	}
	result = totalSizes[0]
	for i := 1; i < len(totalSizes); i++ {
		result = gcd(result, totalSizes[i])
	}
	return result
}

// isValidSetSize - checks whether given count is a valid set size for erasure coding.
var isValidSetSize = func(count uint64) bool {
	return (count >= setSizes[0] && count <= setSizes[len(setSizes)-1])
}

func commonSetDriveCount(divisibleSize uint64, setCounts []uint64) (setSize uint64) {
	// prefers setCounts to be sorted for optimal behavior.
	//已经排序了, 如果小于可选最大值, 直接返回divisibleSize
	if divisibleSize < setCounts[len(setCounts)-1] {
		return divisibleSize
	}

	// Figure out largest value of total_drives_in_erasure_set which results
	// in least number of total_drives/total_drives_erasure_set ratio.
	//找到ec set中最大的磁盘数量, 同时保证集合数量最小
	prevD := divisibleSize / setCounts[0]
	//假如没有省略号时, divisibleSize=4, volume==4, setCounts=[4 8 12 16] prevD=1, 那么pred就是1, setSize=4
	// 如果公约数太大时, 尽可能找符合的比较小的set.
	for _, cnt := range setCounts {
		if divisibleSize%cnt == 0 {
			d := divisibleSize / cnt
			if d <= prevD {
				prevD = d
				setSize = cnt
			}
		}
	}
	return setSize
}

// possibleSetCountsWithSymmetry returns symmetrical setCounts based on the
// input argument patterns, the symmetry calculation is to ensure that
// we also use uniform number of drives common across all ellipses patterns.
func possibleSetCountsWithSymmetry(setCounts []uint64, argPatterns []ellipses.ArgPattern) []uint64 {
	newSetCounts := make(map[uint64]struct{})
	//假如公约数为1(磁盘数为1, 或者两者没有其他公约数.), setCounts为[2,16]
	//计算时, 最好不要拿不太合适的值进行计算, 容易出问题. 就拿比较合理的数据计算. 例如不要出现3/5之类的.
	//假如没有省略号, volume为4, 最大公约数也为4.  此时, setCounts=[4 8 12 16], argPatterns=nil
	for _, ss := range setCounts {
		var symmetry bool
		//如果可能的最大公约数倍数. 最大公约数 * 1/2/3倍
		//再检查一下, 是否符合. 如果最大公约数为1, 反而会有问题.
		for _, argPattern := range argPatterns {
			for _, p := range argPattern {
				if uint64(len(p.Seq)) > ss {
					symmetry = uint64(len(p.Seq))%ss == 0
				} else {
					symmetry = ss%uint64(len(p.Seq)) == 0
				}
			}
		}
		// With no arg patterns, it is expected that user knows
		// the right symmetry, so either ellipses patterns are
		// provided (recommended) or no ellipses patterns.
		//argPatterns == nil 对应 没有省略号的场景.
		if _, ok := newSetCounts[ss]; !ok && (symmetry || argPatterns == nil) {
			newSetCounts[ss] = struct{}{}
		}
	}

	setCounts = []uint64{}
	for setCount := range newSetCounts {
		setCounts = append(setCounts, setCount)
	}

	// Not necessarily needed but it ensures to the readers
	// eyes that we prefer a sorted setCount slice for the
	// subsequent function to figure out the right common
	// divisor, it avoids loops.
	//排序.
	sort.Slice(setCounts, func(i, j int) bool {
		return setCounts[i] < setCounts[j]
	})

	//假如没有省略号, volume为4, 此时, 返回也是 setCounts=[4 8 12 16]
	return setCounts
}

// getSetIndexes returns list of indexes which provides the set size
// on each index, this function also determines the final set size
// The final set size has the affinity towards choosing smaller
// indexes (total sets)
func getSetIndexes(args []string, totalSizes []uint64, customSetDriveCount uint64, argPatterns []ellipses.ArgPattern) (setIndexes [][]uint64, err error) {
	if len(totalSizes) == 0 || len(args) == 0 {
		return nil, errInvalidArgument
	}

	setIndexes = make([][]uint64, len(totalSizes))
	for _, totalSize := range totalSizes {
		// Check if totalSize has minimum range upto setSize
		if totalSize < setSizes[0] || totalSize < customSetDriveCount {
			msg := fmt.Sprintf("Incorrect number of endpoints provided %s", args)
			return nil, config.ErrInvalidNumberOfErasureEndpoints(nil).Msg(msg)
		}
	}

	//totalSizes是每个arg上磁盘的数量
	//找最大公约数. 所有args都能整除的.
	// 这里的一个arg应该就是一个serverPool.
	//每个serverPool 会有节点数量, 和节点磁盘数量, 节点数量*节点磁盘数量=serverPool 总磁盘数.
	//也可能没有, 就是1.
	//这里是可以找所有serverPool中的最大公约数. 但是实际传入时, 对于有省略号的参数, 会一个个传入.
	//所以这里的totalSizes大小还是1.
	commonSize := getDivisibleSize(totalSizes)
	//所有arg中最大公约数
	possibleSetCounts := func(setSize uint64) (ss []uint64) {
		for _, s := range setSizes {
			if setSize%s == 0 {
				ss = append(ss, s)
			}
		}
		return ss
	}

	//获取可能的ec set数量. 获取的就是commonSize的倍数. 最大公约数的倍数
	//如果最大公约数为1, 则需要过滤setCounts. 下面 possibleSetCountsWithSymmetry
	setCounts := possibleSetCounts(commonSize)
	if len(setCounts) == 0 {
		msg := fmt.Sprintf("Incorrect number of endpoints provided %s, number of drives %d is not divisible by any supported erasure set sizes %d", args, commonSize, setSizes)
		return nil, config.ErrInvalidNumberOfErasureEndpoints(nil).Msg(msg)
	}

	var setSize uint64
	// Custom set drive count allows to override automatic distribution.
	// only meant if you want to further optimize drive distribution.
	//如果自定义了ec set数量, 则检查是否合法.
	if customSetDriveCount > 0 {
		msg := fmt.Sprintf("Invalid set drive count. Acceptable values for %d number drives are %d", commonSize, setCounts)
		var found bool
		for _, ss := range setCounts {
			if ss == customSetDriveCount {
				found = true
			}
		}
		//如果在可能的set集合中找不到 自定义的ec set的数量, 则返回错误, 不合法.
		if !found {
			return nil, config.ErrInvalidErasureSetSize(nil).Msg(msg)
		}

		// No automatic symmetry calculation expected, user is on their own
		setSize = customSetDriveCount
		globalCustomErasureDriveCount = true
	} else {
		//如果没有自定义
		// Returns possible set counts with symmetry.
		//setCounts 最大公约数的倍数(所有serverPool的)
		//possibleSetCountsWithSymmetry 再次检查了每个serverPool上的节点数和磁盘数余除为0
		setCounts = possibleSetCountsWithSymmetry(setCounts, argPatterns)

		if len(setCounts) == 0 {
			msg := fmt.Sprintf("No symmetric distribution detected with input endpoints provided %s, drives %d cannot be spread symmetrically by any supported erasure set sizes %d", args, commonSize, setSizes)
			return nil, config.ErrInvalidNumberOfErasureEndpoints(nil).Msg(msg)
		}

		// Final set size with all the symmetry accounted for.
		setSize = commonSetDriveCount(commonSize, setCounts)
	}

	// Check whether setSize is with the supported range.
	if !isValidSetSize(setSize) {
		msg := fmt.Sprintf("Incorrect number of endpoints provided %s, number of drives %d is not divisible by any supported erasure set sizes %d", args, commonSize, setSizes)
		return nil, config.ErrInvalidNumberOfErasureEndpoints(nil).Msg(msg)
	}

	//最后一步, 这里是做什么呢?
	//setSize是每个ec set的大小.
	//totalSizes[i]/setSize 就是有多少个set.
	//setIndexes[i] 就是serverPool上有多少个set, 每个set的大小.
	//这里使用数组来表示每个set大小. 不应该直接作为一个属性么?
	//尽量使用同一个ec set大小.
	//实际上带省略号的 多个serverPool不会同时传入, 只会一个个传入,最后计算的还是一个serverPool上的.
	//为什么, 也很简单, 防止后续划分ep时, 划分到其他serverPool的ep, 增加处理复杂度.
	for i := range totalSizes {
		for j := uint64(0); j < totalSizes[i]/setSize; j++ {
			setIndexes[i] = append(setIndexes[i], setSize)
		}
	}

	return setIndexes, nil
}

// Returns all the expanded endpoints, each argument is expanded separately.
func (s endpointSet) getEndpoints() (endpoints []string) {
	if len(s.endpoints) != 0 {
		return s.endpoints
	}
	for _, argPattern := range s.argPatterns {
		for _, lbls := range argPattern.Expand() {
			endpoints = append(endpoints, strings.Join(lbls, ""))
		}
	}
	s.endpoints = endpoints
	return endpoints
}

// Get returns the sets representation of the endpoints
// this function also intelligently decides on what will
// be the right set size etc.
func (s endpointSet) Get() (sets [][]string) {
	k := uint64(0)
	endpoints := s.getEndpoints()
	for i := range s.setIndexes {
		for j := range s.setIndexes[i] {
			sets = append(sets, endpoints[k:s.setIndexes[i][j]+k])
			k = s.setIndexes[i][j] + k
		}
	}

	return sets
}

// Return the total size for each argument patterns.
func getTotalSizes(argPatterns []ellipses.ArgPattern) []uint64 {
	var totalSizes []uint64
	for _, argPattern := range argPatterns {
		var totalSize uint64 = 1
		for _, p := range argPattern {
			totalSize *= uint64(len(p.Seq))
		}
		totalSizes = append(totalSizes, totalSize)
	}
	return totalSizes
}

// Parses all arguments and returns an endpointSet which is a collection
// of endpoints following the ellipses pattern, this is what is used
// by the object layer for initializing itself.
func parseEndpointSet(customSetDriveCount uint64, args ...string) (ep endpointSet, err error) {
	argPatterns := make([]ellipses.ArgPattern, len(args))
	for i, arg := range args {
		patterns, perr := ellipses.FindEllipsesPatterns(arg)
		if perr != nil {
			return endpointSet{}, config.ErrInvalidErasureEndpoints(nil).Msg(perr.Error())
		}
		argPatterns[i] = patterns
	}

	//getTotalSizes 返回每个arg(volumes)的磁盘数量
	//ep.setIndexes 二维数组, serverPool, serverPool里面每个 ec set的大小.
	//虽然这里可以进行多个serverpool 判断处理, 但是上层传入的时候, 还是一个个serverPool, 传入的.
	ep.setIndexes, err = getSetIndexes(args, getTotalSizes(argPatterns), customSetDriveCount, argPatterns)
	if err != nil {
		return endpointSet{}, config.ErrInvalidErasureEndpoints(nil).Msg(err.Error())
	}

	ep.argPatterns = argPatterns

	return ep, nil
}

// GetAllSets - parses all ellipses input arguments, expands them into
// corresponding list of endpoints chunked evenly in accordance with a
// specific set size.
// For example: {1...64} is divided into 4 sets each of size 16.
// This applies to even distributed setup syntax as well.
//计算集合数量和大小, 然后进行划分ep. 返回划分好的ep
func GetAllSets(args ...string) ([][]string, error) {
	var customSetDriveCount uint64
	//minio MINIO_ERASURE_SET_DRIVE_COUNT 每个ec集合中有几个磁盘数量, 可以通过环境变量指定.
	//
	if v := env.Get(EnvErasureSetDriveCount, ""); v != "" {
		driveCount, err := strconv.Atoi(v)
		if err != nil {
			return nil, config.ErrInvalidErasureSetSize(err)
		}
		customSetDriveCount = uint64(driveCount)
	}

	var setArgs [][]string
	//如果传入的args没有省略号
	if !ellipses.HasEllipses(args...) {
		var setIndexes [][]uint64
		// Check if we have more one args.
		//有多个volume
		if len(args) > 1 {
			var err error
			//没有省略号时, 多个volume, totalSizes就是一个元素的数组, 数组长度为volume数量.
			//假如有4个volume
			//commonSize = volume数量.
			//setCounts = 4 8 12 16
			//没有省略号时, 只有一个serverPool.
			//setIndexes是一个二维数组,表示每个serverPool所有的set, 以及set对应的集合大小是多少.
			//当然当前场景下setIndexes只有一个0值对应的数组.
			setIndexes, err = getSetIndexes(args, []uint64{uint64(len(args))}, customSetDriveCount, nil)
			if err != nil {
				return nil, err
			}
		} else {
			// We are in FS setup, proceed forward.
			setIndexes = [][]uint64{{uint64(len(args))}}
		}
		s := endpointSet{
			endpoints:  args,
			setIndexes: setIndexes,
		}
		setArgs = s.Get()
	} else {
		//传入的args有省略号时
		//这里面对volume ec分多少个纠删集合 ec set.
		//ep.setIndexes 二维数组, serverPool, serverPool里面每个 ec set的大小.
		//就是这里的s
		s, err := parseEndpointSet(customSetDriveCount, args...)
		if err != nil {
			return nil, err
		}
		//根据ec集合的大小进行划分不同的set上对应哪些endpoint
		setArgs = s.Get()
	}

	//保证ep不重复.
	uniqueArgs := set.NewStringSet()
	for _, sargs := range setArgs {
		for _, arg := range sargs {
			if uniqueArgs.Contains(arg) {
				return nil, config.ErrInvalidErasureEndpoints(nil).Msg(fmt.Sprintf("Input args (%s) has duplicate ellipses", args))
			}
			uniqueArgs.Add(arg)
		}
	}

	//返回划分好的ep
	return setArgs, nil
}

// Override set drive count for manual distribution.
const (
	EnvErasureSetDriveCount = "MINIO_ERASURE_SET_DRIVE_COUNT"
)

var globalCustomErasureDriveCount = false

// CreateServerEndpoints - validates and creates new endpoints from input args, supports
// both ellipses and without ellipses transparently.
//args传递的是MINIO_VOLUMES
//这里首先划分set, 然后划分每个set中的ep, 然后创建ep, 确定setuptype类型.
func createServerEndpoints(serverAddr string, args ...string) (
	endpointServerPools EndpointServerPools, setupType SetupType, err error,
) {
	if len(args) == 0 {
		return nil, -1, errInvalidArgument
	}

	ok := true
	//查看每一个参数是否有省略号
	//查看volume中有没有省略号.
	for _, arg := range args {
		ok = ok && !ellipses.HasEllipses(arg)
	}

	// None of the args have ellipses use the old style.
	//如果有没有省略号, 直接走旧流程
	//就一个serverPool, 每个serverPool下有多个endpoint, 一个挂载点一个endpoint.
	if ok {
		//计算集合数量和大小, 然后进行划分ep. 返回划分好的ep. 顺序划分的.
		setArgs, err := GetAllSets(args...)
		if err != nil {
			return nil, -1, err
		}
		//CreateEndpoints创建endPoint, 同时根据ep 返回不同的SetupType.
		//创建ep时, 如果是路径则是isLocal=true, 如果是url, 且为当前节点, 也是isLocal=true
		endpointList, newSetupType, err := CreateEndpoints(serverAddr, false, setArgs...)
		if err != nil {
			return nil, -1, err
		}
		endpointServerPools = append(endpointServerPools, PoolEndpoints{
			Legacy:       true,
			SetCount:     len(setArgs),
			DrivesPerSet: len(setArgs[0]),
			Endpoints:    endpointList,
			CmdLine:      strings.Join(args, " "),
		})
		setupType = newSetupType
		return endpointServerPools, setupType, nil
	}

	//有省略号, 走后续流程.
	var foundPrevLocal bool
	for _, arg := range args {
		//如果有多个volume, 但是当前volume 并没有省略号, 则为错误.
		if !ellipses.HasEllipses(arg) && len(args) > 1 {
			// TODO: support SNSD deployments to be decommissioned in future
			return nil, -1, fmt.Errorf("all args must have ellipses for pool expansion (%w) args: %s", errInvalidArgument, args)
		}
		//每个serverPool都有自己的ec set
		//划分好ec set集合大小 划分好每个set对应的ep.
		setArgs, err := GetAllSets(arg)
		if err != nil {
			return nil, -1, err
		}

		//获取setup类型.
		endpointList, gotSetupType, err := CreateEndpoints(serverAddr, foundPrevLocal, setArgs...)
		if err != nil {
			return nil, -1, err
		}
		if err = endpointServerPools.Add(PoolEndpoints{
			//多少个set
			SetCount:     len(setArgs),
			//每个set上有多少个磁盘.
			//为啥是0呢? 首先, 对于多个serverPool, 只会传入1个. 第二即使是多个也是相同的大小.
			//setArgs[0]里面存放的是set 0上的所有ep. 划分好的ep.
			DrivesPerSet: len(setArgs[0]),
			Endpoints:    endpointList,
			CmdLine:      arg,
		}); err != nil {
			return nil, -1, err
		}
		foundPrevLocal = endpointList.atleastOneEndpointLocal()
		if setupType == UnknownSetupType {
			setupType = gotSetupType
		}
		if setupType == ErasureSetupType && gotSetupType == DistErasureSetupType {
			setupType = DistErasureSetupType
		}
	}

	return endpointServerPools, setupType, nil
}
