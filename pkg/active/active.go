package active

import (
	"context"
	"github.com/hary654321/Starmap/pkg/resolve"
	"github.com/hary654321/Starmap/pkg/util"
	"github.com/projectdiscovery/gologger"
)

func Enum(domain string, uniqueMap map[string]resolve.HostEntry, silent bool, fileName string, level int, levelDict string, resolvers []string, wildcardIPs map[string]struct{}, maxIPs int) (map[string]resolve.HostEntry, map[string]struct{}) {
	gologger.Info().Msgf("Start DNS blasting of %s", domain)
	var levelDomains []string
	if levelDict != "" {
		dl, err := util.LinesInFile(levelDict)
		if err != nil {
			gologger.Fatal().Msgf("读取domain文件失败:%s,请检查--level-dict参数\n", err.Error())
		}
		levelDomains = dl
	} else {
		levelDomains = GetDefaultSubNextData()
	}

	opt := &Options{
		Rate:         Band2Rate("2m"),
		Domain:       domain,
		FileName:     fileName,
		Resolvers:    resolvers,
		Output:       "",
		Silent:       silent,
		WildcardIPs:  wildcardIPs,
		MaxIPs:       maxIPs,
		TimeOut:      5,
		Retry:        6,
		Level:        level,        // 枚举几级域名，默认为2，二级域名,
		LevelDomains: levelDomains, // 枚举多级域名的字典文件，当level大于2时候使用，不填则会默认
		Method:       "enum",
	}

	ctx := context.Background()

	r, err := New(opt)

	if err != nil {
		gologger.Fatal().Msgf("%s", err)
	}

	enumMap, wildcardIPs := r.RunEnumeration(uniqueMap, ctx)

	r.Close()

	return enumMap, wildcardIPs
}

func Verify(uniqueMap map[string]resolve.HostEntry, silent bool, resolvers []string, wildcardIPs map[string]struct{}, maxIPs int) (map[string]resolve.HostEntry, map[string]struct{}, []string) {
	gologger.Info().Msgf("Start to verify the collected sub domain name results, a total of %d", len(uniqueMap))

	opt := &Options{
		Rate:        Band2Rate("2m"),
		Domain:      "",
		UniqueMap:   uniqueMap,
		Resolvers:   resolvers,
		Output:      "",
		Silent:      silent,
		WildcardIPs: wildcardIPs,
		MaxIPs:      maxIPs,
		TimeOut:     5,
		Retry:       6,
		Method:      "verify",
	}
	ctx := context.Background()

	r, err := New(opt)
	if err != nil {
		gologger.Fatal().Msgf("%s", err)
	}

	AuniqueMap, wildcardIPs, unanswers := r.RunEnumerationVerify(ctx)

	r.Close()

	return AuniqueMap, wildcardIPs, unanswers
}
