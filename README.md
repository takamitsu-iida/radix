# radix

key-valueペアを保持するデータ構造です。

キーに合致する値だけでなく、キーに対してロンゲストマッチ方式で情報を取り出せます。

## 実装メモ

自分でもすぐに忘れてしまうのでメモを残します。

radix treeの構造自体はWikipediaの説明にあるものとだいたい同じです。

english
https://en.wikipedia.org/wiki/Radix_tree

japanese
https://ja.wikipedia.org/wiki/%E5%9F%BA%E6%95%B0%E6%9C%A8


<br><br>

### Tree

radix treeの本体です。

```go
type Tree struct {
	root *node
	size int
}
```

<dl>
  <dt>root</dt>  <dd>ルートノードです。ルートノードからはエッジが伸びていきます。ルートノードがリーフを保持することはありません。</dd>
  <dt>size</dt>  <dd>Tree内に格納されているkey-valueペア（リーフ）の数です。</dd>
</dl>

<br><br>

### leaf

key-valueペアを保持する構造体です。leafを作成したらTreeのsizeを+1し、削除したらsizeを-1します。

```go
type leaf struct {
	key string
	val interface{}
}
```

<dl>
  <dt>key</dt>  <dd>文字列です。</dd>
  <dt>val</dt>  <dd>インタフェースとして定義していますので、適宜キャストが必要です。</dd>
</dl>

<br><br>

### node

ツリーを構成するノードです。

```go
type node struct {
	leaf *leaf
	prefixes []rune
	edges []edge
}
```

<dl>
  <dt>leaf</dt>  <dd>ノードが値を保持する場合は、リーフを作成し、そこへのポインタを保持します。リーフを持たない場合、leafはnilになります。</dd>
  <dt>prefixes</dt>  <dd>このノードにたどり着くまでの共通部分を除いたユニーク部分です。keyがUTF-8の場合を想定してruneのスライスです。</dd>
  <dt>edges</dt>  <dd>このノードから分岐していくエッジを格納するスライスです。常に辞書順にソートされています。</dd>
</dl>

<br><br>

### edge

ツリーを構成するエッジです。

```go
type edge struct {
	label rune
	node  *node
}
```

<dl>
  <dt>label</dt>  <dd>このエッジを識別するrune型の単一文字です。ノードから分岐するエッジは常に辞書順にソートされているので、この文字を使って探し当てるのは簡単です。</dd>
  <dt>node</dt>  <dd>そのエッジの先にいるノード、つまり子ノードです。</dd>
</dl>

<br><br>

### ツリー構造のイメージ

ツリー構造はこのような形をしています。

`-(edge)-` はエッジを、`[node]` はノードを意味します。

```txt
[node]--+--(edge)--[node]
        +--(edge)--[node]
        +--(edge)--[node]
```

例えば、

- romance
- romanus
- romulus
- rubens

をキーとして格納することを考えます。

rommanceの先頭はrですので、ルートノードからラベルrを持ったエッジを探しますが、まだ何も格納されていないので見つかりません。
ラベルrを持ったエッジを新規に作成してルートノードにぶら下げます。
子ノード[rommance]も新規に作成してエッジにぶら下げます。
キーをromanceとしたkey-valueペアをリーフとして作成しノードに保持します。

この処理でこのようなツリー構造になりました。

```txt
root-(r)-[romance]
```

次にromanusをツリーに加えるとします。

romanusの先頭文字はrですので、ラベルrを持ったエッジを探します。
今度はすでに存在しますので、その先の子ノード[romance]に到達します。

[romance]が持つプレフィクスと探索中のromanusを比較します。共通部分を取り出すとromanです。
すでにある[romance]を、共通部分である[roman]と、ユニークな部分である[ce]に分割します。
探索中のromanusは共通部分を削除し、ユニークな部分だけを取り出して[us]というノードを作ります。

この処理でこのようなツリー構造になりました。[roman]から二つに分岐しています。

```txt
root-(r)-[roman]-+-(c)-[ce]
                 |
                 +-(u)-[us]
```

次にromulusを加えるとします。

romulusはrで始まりますので、ラベルrを持ったエッジを探し、その先の子ノード[roman]を見つけます。
[roman]とromulusで共通部分を取り出すとromになります。
すでにある[roman]を二つに分割し、共通部分の[rom]とユニーク部分の[an]に分けます。
探索中のromulusはユニークな部分を取り出して[ulus]というノードを作ります。

この処理でこのようなツリー構造になりました。[rom]から二つに分岐し、[an]から二つに分岐しています。

```txt
root-(r)-[rom]-+-(a)-[an]-+-(c)-[ce]
               |          |
               |          +-(u)-[us]
               |
               +-(u)-[ulus]
```

次にrubensを加えるとします。

rubensの先頭文字はrですので、ラベルrを持ったエッジを探します。
この場合はすでに存在しますので、その先の子ノード[rom]に到達します。
[rom]とrubensで共通部分を取り出すとrになります。
すでにある[rom]を二つに分割し、共通部分の[r]とユニーク部分の[om]に分けます。
探索中のrubensはユニークな部分を取り出して[ubens]というノードを作ります。

最終的にはこのようなツリー構造になります。

```txt
root-(r)-[r]-+-(o)-[om]-+-(a)-[an]-+-(c)-[ce]
             |          |          |
             |          |          +-(u)-[us]
             |          |
             |          +-(u)-[ulus]
             |
             +-(u)-[ubens]
```

rootノードからはエッジが一つだけ伸びていてラベルは(r)です。

その子ノードは[r]で、プレフィクスrを持ちます。
[r]は中間ノードなのでリーフを持ちません。
[r]からは2本のエッジが伸びています。一つはラベル(o)を、もう一つはラベル(u)を持ちます。

ラベル(o)のエッジの先には子ノード[om]がいます。これも中間ノードなのでリーフを持ちません。
[om]からは2本のエッジが伸びています。一つはラベル(a)を、もう一つは(u)を持ちます。

ラベル(a)のエッジの先には子ノード[an]がいます。
[an]は中間ノードなのでリーフを持ちません。
[an]からは2本のエッジが伸びています。一つはラベル(c)を、もう一つは(u)を持ちます。

ラベル(c)の先には子ノード[ce]がいます。これはリーフでkey-valueペアを持ち、キーはrommanceです。

このようにツリーを探索しながらキーを挿入するわけです。

<br><br>

### 探索の実装

ツリーに格納しているkey-valueペアを探し当てるには、まずkey文字列をruneのスライスに変換して、探索キーを生成します。
探索キーと、ツリー内のノードのプレフィクスを比較して、共通する部分を探索キーから削りながらツリーをたどっていきます。
探索キーの長さがゼロになったら、格納しているkey-valueペアのノードにたどり着いたことになります。

たとえば、前述のツリーのなかにあるromanceを探索するとします。

ノード[r]に到達すると探索キーはrを削除してomanceになります。ノード[om]に到達すると探索キーはomを削除してanceになります。ノード[an]に到達すると探索キーは[ce]になります。ノード[ce]に到達すると探索キーはなくなり長さゼロになります。この状態になれば、リーフrommanceを持つノードに到達したことになります。

探索キーの長さがゼロではないのに次の子ノードが見つからない場合、それは探しているkey-valueペアのキーがツリーに格納されていないということです。

もし見つからなくても、最後に一致したノードの情報を返却するようにすれば、ロンゲストマッチでの探索になります。

<br><br>

### ルーティングテーブルの検索

IPの経路情報はロンゲストマッチ方式で転送先のゲートウェイが選ばれます。

たとえば、10.0.0.0/8はgig1に転送、10.0.0.0/16はgig2に転送、10.0.0.0/24はgig3に転送するとします。
宛先10.0.0.1のパケットが転送されるべきゲートウェイを考えます。
この場合/8と/16と/24すべてのエントリに合致するのですが、一番長いマスク長は/24なので、gig3が選ばれます。

この検索を容易に実現できるのがradix treeの特徴です。

実際にradix treeに経路情報を格納するときには、プレフィクス情報をビット列に変換して格納すると簡単になります。

10.0.0.0/8をビット表現すると`00001010`となります。/8なので32ビットのうち先頭8ビットだけを使います。

10.0.0.0/16をビット表現すると`0000101000000000`となります。/16なので32ビットのうち先頭16ビットだけを使います。

10.0.0.0/24をビット表現すると`000010100000000000000000`となります。/24なので32ビットのうち先頭24ビットだけを使います。

宛先10.0.0.1をビット表現すると`00001010000000000000000000000001`になりますので、これを探索キーとしてツリーを探せば/24のエントリが最も深いところまでたどり着くことがわかるでしょう。最後に見つけたリーフノードを返却するのがロンゲストマッチ方式、ということです。

radix_test.goにこのテスト（↓）を書きました。もちろん期待通りにPASSします。

```go
func TestIP(t *testing.T) {
	// routing table
	routes := []struct {
		prefix  string
		gateway string
	}{
		{"10.0.0.0/8", "gig1"},
		{"10.0.0.0/16", "gig2"},
		{"10.0.0.0/24", "gig3"},
		{"192.168.0.0/24", "gig4"},
		{"192.168.0.128/25", "gig5"},
	}

	r := New()

	// convert prefix to a bit string, then insert into radix tree
	for _, route := range routes {
		addr, masklen, err := cidrToBinaryString(route.prefix)
		if err != nil {
			t.Fatalf("failed to convert string: %v", route.prefix)
		}
		addr = addr[:masklen]
		r.Insert(addr, route.gateway)
	}

	tests := []struct {
		destination string
		expected    string
	}{
		{"10.0.0.1", "gig3"},
		{"10.0.1.1", "gig2"},
		{"10.1.1.1", "gig1"},
		{"192.168.0.1", "gig4"},
		{"192.168.0.129", "gig5"},
	}

	for _, test := range tests {
		addr, err := addrToBinaryString(test.destination)
		if err != nil {
			t.Fatal("failed to convert string", err)
		}

		_, v, found := r.LongestMatch(addr)
		if found == false {
			t.Fatalf("key not found: %v", test.destination)
		}
		if test.expected != v {
			t.Fatalf("expected: %v, got: %v", test.expected, v)
		}
	}
}
```

<br><br>

### ドメインリストの検索

ドメイン名でフィルタリングする場合もロンゲストマッチ方式での検索が使われます。
たとえばこの二つ。

- "teams.microsoft.com"
- "microsoft.com"

teams.microsoft.comの方が長いドメイン名を持っていますので、こちらを優先的に処理すべきです。

また、

- *.microsoft.com"

のように先頭にアスタリスクをつけて任意のサブドメイン名を表すこともよくあります。
URLは文字列の先頭側がより詳細な情報を表していますので、絞り込みはやりづらい形です。
ドメイン名をキーとして格納するときには順番をひっくり返して表現すると検索が容易になります。

`teams.microsoft.com` であれば `com.microsoft.teams` のように変換します。
`*.microsoft.com` であれば `com.microsoft.*` のように最後にアスタリスクがきます。
この場合、最後のアスタリスクを削除して `com.microsoft.` としてキーを表現すればよいでしょう。

radix_test.goにこのテスト（↓）を書きました。もちろん期待通りにPASSします。

```go
func TestDomain(t *testing.T) {
	// domain list
	domains := []string{
		"teams.microsoft.com",
		"microsoft.com",
		"*.teams.microsoft.com",
		"*.microsoft.com",
	}

	r := New()

	for _, domain := range domains {
		reversed := reverseUrl(domain)
		if reversed[len(reversed)-1] == '*' {
			reversed = reversed[:len(reversed)-1]
		}
		r.Insert(reversed, domain)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"teams.microsoft.com", "teams.microsoft.com"},
		{"a.teams.microsoft.com", "*.teams.microsoft.com"},
		{"a.b.teams.microsoft.com", "*.teams.microsoft.com"},
		{"abc.microsoft.com", "*.microsoft.com"},
	}

	for _, test := range tests {
		reveresed := reverseUrl(test.input)
		_, v, found := r.LongestMatch(reveresed)
		if found == false {
			t.Fatalf("key not found: %v", test.input)
		}
		if v != test.expected {
			t.Fatalf("input: %v, expected: %v, got: %v", test.input, test.expected, v)
		}
	}
}
```

なお、URLの途中にアスタリスクが来る場合には対応できません。
これに対応するにはもうちょっと工夫が必要です。

<!--
どう工夫するか

ノードは、自身のプレフィクスにワイルドカードを含むかどうかを識別できるようにする。

キー a*z を初めて登録するときには、このようにワイルドカードの前で区切って二つに分割する。

root-(a)-[a]-(*)-[*z, wc=true]

次にabcをキーとして登録することを考える。
ノード[a]を訪問した時点で共通部を削除するので探索キーは[bc]になっている。
bcに対して*zで検索しても該当しないので、共通部分は存在しない、と考える。
エッジの順序として、*が最後になるように配慮する。

root-(a)-[a]-+-(b)-[bc]
             |
             +-(*)-[*z, wc=true]

次にabzをキーとして登録することを考える
ノード[a]を訪問した時点で探索キーはbzになっている。
エッジ(b)をたどって[bc]を訪問する。
bcとbzで共通部はbなので、[bc]は[b]と[c]の二つに分割する。
探索キーは共通部を削除してzになるので、ノード[z]を生成する。

root-(a)-[a]-+-(b)-[b]-+-(c)-[c]
             |         |
             |         +-(z)-[z]
             +-(*)-[*z, wc=true]

abzはa*zにも該当するが、より詳しい情報である[a]-[bc]を辿るのでワイルドカードの検索にはヒットしない。

次にaxzをキーとして登録することを考える。
ラベル(a)を辿ってノード[a]に到達する。
ノード[a]にはエッジ(b)と(*)があり、探索キーxzは(*)にのみ一致する。
xzに対して*zを検索すると該当するため、ここを起点とした新しいツリーを生成する。
エッジ(x)を伸ばしてノード[xz]をぶら下げる

root-(a)-[a]-+-(b)-[b]-+-(c)-[c]
             |         |
             |         +-(z)-[z]
             |
             +-(*)-[*z, wc=true]-(x)-[xz]

axyzを登録する。

root-(a)-[a]-+-(b)-[b]-+-(c)-[c]
             |         |
             |         +-(z)-[z]
             |
             +-(*)-[*z, wc=true]-(x)-[x]-+-(z)-[z]
                                         |
                                         +-(y)-[yz]


この状態でaazを探してみる。
ラベル(a)を辿ってノード[a]に到着。ラベルaのエッジは存在しないので、[*z]に到着。
ここを起点にazを探す。
見つからない。
[*z]のリーフを返却して終了。

というように、*で始まるノードを起点としたツリーを作っていけばよい。

と思っているのだが、どうだろう？
-->