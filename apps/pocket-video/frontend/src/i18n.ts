export type Locale = 'en' | 'zh-CN' | 'ja';
export type Copy = {
 home:string; favorites:string; videos:string; queue:string; settings:string; search:string; hero:string; subtitle:string; browse:string;
 saved:string; recent:string; all:string; download:string; empty:string; preferences:string; language:string; theme:string; system:string;
 light:string; dark:string; cache:string; folder:string; frequency:string; automatic:string; proxy:string; save:string; savedmsg:string;
 sync:string; offline:string; back:string; library:string; source:string; account:string; noVideo:string; escape:string;
 libraryCaption:string; personalEyebrow:string; collectionEyebrow:string; favoriteEyebrow:string; transferEyebrow:string; makeItYoursEyebrow:string;
 status:string; audio:string; menu:string; profile:string; brandVersion:string; collectionIndex:string; videosUnit:string; available:string;
 synced:string; today:string; languageHint:string; themeHint:string; cacheHint:string; folderHint:string; frequencyHint:string; automaticHint:string;
 proxyHint:string; queued:string; running:string; paused:string; done:string; failed:string; videoKind:string; audioKind:string;
};

export const locales: Locale[] = ['en', 'zh-CN', 'ja'];

export const copy: Record<Locale, Copy> = {
 en:{
  home:'Home',favorites:'Favorites',videos:'Videos',queue:'Download queue',settings:'Settings',search:'Search your library...',
  hero:'Your favorites,\nanywhere.',subtitle:'A quieter place for the videos you want to keep.',browse:'Browse favorites',saved:'Saved offline',
  recent:'Recently added',all:'All videos',download:'Download',empty:'No downloads yet',preferences:'Preferences',language:'Language',
  theme:'Appearance',system:'System',light:'Light',dark:'Dark',cache:'Cache location',folder:'Download location',
  frequency:'Sync every (minutes)',automatic:'Sync automatically',proxy:'Proxy URL',save:'Save changes',savedmsg:'Settings saved',
  sync:'Sync favorites',offline:'Offline ready',back:'Back',library:'A small library, just for you.',
  source:'Development mode · deterministic demo data',account:'DEMO ACCOUNT',noVideo:'Download this video to watch offline.',
  escape:'Keyboard & gamepad: arrows move, Enter confirms, Esc goes back.',libraryCaption:'LIBRARY',
  personalEyebrow:'YOUR PERSONAL OFFLINE LIBRARY',collectionEyebrow:'YOUR COLLECTION',favoriteEyebrow:'FAVORITE · POCKETJS',
  transferEyebrow:'TRANSFER ACTIVITY',makeItYoursEyebrow:'MAKE IT YOURS',status:'STATUS',audio:'Audio',menu:'Menu',profile:'Profile',
  brandVersion:'V1.0.0',collectionIndex:'COLLECTION · 01',videosUnit:'VIDEOS',available:'AVAILABLE',synced:'SYNCED',today:'TODAY',
  languageHint:'Choose your preferred language',themeHint:'Light, dark, or match your device',cacheHint:'Temporary artwork and metadata',
  folderHint:'Keep your downloaded media here',frequencyHint:'How often to check for new favorites',
  automaticHint:'Keep your collection up to date',proxyHint:'Optional HTTP or SOCKS proxy',
  queued:'QUEUED',running:'RUNNING',paused:'PAUSED',done:'DONE',failed:'FAILED',videoKind:'VIDEO',audioKind:'AUDIO'
 },
 'zh-CN':{
  home:'首页',favorites:'收藏夹',videos:'视频',queue:'下载队列',settings:'设置',search:'搜索收藏内容...',
  hero:'喜欢的视频，\n随时观看。',subtitle:'把想留住的视频，安静地收藏起来。',browse:'浏览收藏夹',saved:'已离线保存',
  recent:'最近添加',all:'全部视频',download:'下载',empty:'暂无下载任务',preferences:'偏好设置',language:'语言',
  theme:'外观主题',system:'跟随系统',light:'浅色',dark:'深色',cache:'缓存目录',folder:'下载目录',
  frequency:'同步间隔（分钟）',automatic:'自动同步',proxy:'代理地址',save:'保存设置',savedmsg:'设置已保存',
  sync:'同步收藏夹',offline:'可离线观看',back:'返回',library:'属于你的私人影像库。',
  source:'开发模式 · 确定性演示数据',account:'演示账号',noVideo:'下载此视频后即可离线观看。',
  escape:'键盘与手柄：方向键移动，Enter 确认，Esc 返回。',libraryCaption:'资料库',
  personalEyebrow:'你的个人离线资料库',collectionEyebrow:'你的收藏',favoriteEyebrow:'收藏 · POCKETJS',
  transferEyebrow:'传输活动',makeItYoursEyebrow:'让它成为你的',status:'状态',audio:'音频',menu:'菜单',profile:'个人资料',
  brandVersion:'V1.0.0',collectionIndex:'合集 · 01',videosUnit:'个视频',available:'可用',synced:'已同步',today:'今天',
  languageHint:'选择你偏好的语言',themeHint:'浅色、深色或跟随设备',cacheHint:'用于缓存封面和元数据',
  folderHint:'将下载的媒体保存在这里',frequencyHint:'检查新收藏内容的频率',
  automaticHint:'让收藏内容保持最新',proxyHint:'可选的 HTTP 或 SOCKS 代理',
  queued:'排队中',running:'下载中',paused:'已暂停',done:'已完成',failed:'失败',videoKind:'视频',audioKind:'音频'
 },
 ja:{
  home:'ホーム',favorites:'お気に入り',videos:'動画',queue:'ダウンロード',settings:'設定',search:'ライブラリを検索...',
  hero:'好きな動画を、\nいつでも。',subtitle:'残したい動画を、静かな場所に保存。',browse:'お気に入りを見る',saved:'オフライン保存済み',
  recent:'最近追加',all:'すべての動画',download:'ダウンロード',empty:'ダウンロードはありません',preferences:'環境設定',language:'言語',
  theme:'外観',system:'システム',light:'ライト',dark:'ダーク',cache:'キャッシュの場所',folder:'保存先',
  frequency:'同期間隔（分）',automatic:'自動同期',proxy:'プロキシ URL',save:'変更を保存',savedmsg:'設定を保存しました',
  sync:'お気に入りを同期',offline:'オフライン再生可能',back:'戻る',library:'あなただけの小さなライブラリ。',
  source:'開発モード · 決定的なデモデータ',account:'デモアカウント',noVideo:'ダウンロードするとオフラインで視聴できます。',
  escape:'キーボードとゲームパッド：矢印で移動、Enterで決定、Escで戻る。',libraryCaption:'ライブラリ',
  personalEyebrow:'あなたのオフラインライブラリ',collectionEyebrow:'コレクション',favoriteEyebrow:'お気に入り · POCKETJS',
  transferEyebrow:'転送アクティビティ',makeItYoursEyebrow:'自分好みに',status:'ステータス',audio:'音声',menu:'メニュー',profile:'プロフィール',
  brandVersion:'V1.0.0',collectionIndex:'コレクション · 01',videosUnit:'本の動画',available:'利用可能',synced:'同期済み',today:'今日',
  languageHint:'表示言語を選択',themeHint:'ライト、ダーク、またはデバイスに合わせる',cacheHint:'アートワークとメタデータの一時保存先',
  folderHint:'ダウンロードしたメディアの保存先',frequencyHint:'新しいお気に入りを確認する間隔',
  automaticHint:'コレクションを最新に保つ',proxyHint:'任意の HTTP または SOCKS プロキシ',
  queued:'待機中',running:'実行中',paused:'一時停止',done:'完了',failed:'失敗',videoKind:'動画',audioKind:'音声'
 }
};
