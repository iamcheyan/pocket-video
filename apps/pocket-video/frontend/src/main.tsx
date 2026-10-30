import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Home, Library, Search, Download, Settings as SettingsIcon, Play, Pause, Globe, Sun, Moon, Monitor, Check, ChevronRight, Heart, Clock, HardDrive, RefreshCw, ArrowLeft } from 'lucide-react';
import { installNavigation } from './navigation';
import { copy as dict, type Copy, type Locale } from './i18n';
import './style.css';

type Video = { id:string; favorite_id:string; title:string; creator:string; duration:number; has_video:boolean; has_audio:boolean };
type Task = { id:number; video_id:string; kind:string; status:string; progress:number };
type Settings = { language:string; theme:string; cache_dir:string; download_dir:string; sync_minutes:number; auto_sync:boolean; proxy:string; error?:{code:string;message:string} };
function formatDuration(seconds:number){return `${Math.floor(seconds/60)}:${String(seconds%60).padStart(2,'0')}`}

function VideoGrid({videos,open,download,t}:{videos:Video[];open:(video:Video)=>void;download:(video:Video)=>void;t:Copy}){
 return <div className="video-grid">{videos.map((video,index)=><article className="video-card" key={video.id}>
  <button className="thumbnail" onClick={()=>open(video)} aria-label={video.title}>
   <div className={'landscape landscape-'+index%3}><span className="sun"/><div className="mountains"/><div className="lake"/></div>
   <span className="duration">{formatDuration(video.duration)}</span><span className="thumb-play"><Play size={14} fill="currentColor"/></span>
  </button>
  <div className="video-meta"><div><span className="eyebrow">{t.collectionIndex}</span><h3><button onClick={()=>open(video)}>{video.title}</button></h3><p>{video.creator}</p></div>
   <button className="download-icon" aria-label={t.download} onClick={()=>download(video)}>{video.has_video?<Check size={16}/>:<Download size={16}/>}</button>
  </div>
 </article>)}</div>
}

function App(){
 const [lang,setLang]=useState<Locale>('en');
 const [page,setPage]=useState('home');
 const [videos,setVideos]=useState<Video[]>([]);
 const [tasks,setTasks]=useState<Task[]>([]);
 const [selected,setSelected]=useState<Video|null>(null);
 const [query,setQuery]=useState('');
 const [settings,setSettings]=useState<Settings>({language:'en',theme:'system',cache_dir:'',download_dir:'',sync_minutes:60,auto_sync:false,proxy:''});
 const [saved,setSaved]=useState(false);
 const [mobileNav,setMobileNav]=useState(false);
 const t=dict[lang];
 const api=async(path:string,options?:RequestInit)=>{try{const response=await fetch('/api'+path,options);return await response.json()}catch{return {error:{code:'network',message:'PocketVideo service unavailable'}}}};
 const applyTheme=(value:string)=>{document.documentElement.dataset.theme=value==='dark'||(value==='system'&&matchMedia('(prefers-color-scheme: dark)').matches)?'dark':'light'};
 const refresh=async()=>{const [videoResult,taskResult]=await Promise.all([api('/videos?q='+encodeURIComponent(query)),api('/task')]);setVideos(videoResult.items||[]);setTasks(Array.isArray(taskResult)?taskResult:[])};
 useEffect(()=>{api('/settings').then((s:Settings)=>{if(s.error)return;setSettings(s);setLang(s.language as Locale);localStorage.lang=s.language;applyTheme(s.theme)});api('/task').then((x:Task[])=>setTasks(Array.isArray(x)?x:[]))},[]);
 useEffect(()=>{api('/videos?q='+encodeURIComponent(query)).then(x=>setVideos(x.items||[]))},[query]);
 useEffect(()=>{const timer=setInterval(()=>api('/task').then((x:Task[])=>setTasks(Array.isArray(x)?x:[])),1000);return()=>clearInterval(timer)},[]);
 useEffect(()=>installNavigation(
  ()=>{if(mobileNav){setMobileNav(false);return}if(page==='detail'){setPage('videos');return}if(page!=='home'){setPage('home')}},
  ()=>{const focused=document.activeElement;if(focused instanceof HTMLElement&&focused!==document.body)focused.click();else document.querySelector<HTMLElement>('.nav-item.active')?.focus()}
 ),[page,mobileNav]);
 const changeLanguage=async(value:string)=>{setLang(value as Locale);localStorage.lang=value;const next={...settings,language:value};setSettings(next);await api('/settings',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(next)})};
 const chooseTheme=(value:string)=>{setSettings({...settings,theme:value});applyTheme(value)};
 const saveSettings=async()=>{const next={...settings,language:lang};const result=await api('/settings',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(next)});if(result.error)return;setSettings(next);applyTheme(next.theme);localStorage.lang=lang;setSaved(true);setTimeout(()=>setSaved(false),1800)};
 const startDownload=async(video:Video,kind='video')=>{const result=await api('/download',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({video_id:video.id,kind})});if(result.error)return;api('/task').then((x:Task[])=>setTasks(Array.isArray(x)?x:[]));setPage('queue')};
 const openVideo=(video:Video)=>{setSelected(video);setPage('detail')};
 const navigation=[['home',Home],['favorites',Heart],['videos',Library],['queue',Download],['settings',SettingsIcon]] as const;
 const statusKey: Record<string,keyof Copy>={queued:'queued',running:'running',paused:'paused',done:'done',failed:'failed'};
 return <div className="app-shell">
  <aside className={'sidebar '+(mobileNav?'show':'')}>
   <div className="brand"><span className="brand-mark"><Play size={17} fill="currentColor"/></span><span>Pocket<span className="brand-light">Video</span></span><button className="close-nav" aria-label={t.back} onClick={()=>setMobileNav(false)}>×</button></div>
   <div className="nav-caption">{t.libraryCaption}</div>
   <nav>{navigation.map(([key,Icon])=><button className={'nav-item '+(page===key?'active':'')} key={key} onClick={()=>{setPage(key);setMobileNav(false)}}><Icon size={17}/><span>{t[key]}</span>{key==='queue'&&tasks.some(task=>task.status==='running')&&<i/>}</button>)}</nav>
   <div className="sidebar-bottom"><div className="account-card"><div className="avatar">PV</div><div><b>{t.account}</b><small>{t.source.split(' · ')[0]}</small></div><span className="green-dot"/></div><div className="side-foot">POCKETVIDEO <span>{t.brandVersion}</span></div></div>
  </aside>
  <main className="main">
   <header className="topbar"><button className="mobile-menu" aria-label={t.menu} onClick={()=>setMobileNav(!mobileNav)}>☰</button><div className="breadcrumb">PocketVideo <ChevronRight size={13}/><span>{selected&&page==='detail'?selected.title:t[page as keyof Copy]||t.home}</span></div><div className="top-actions"><label className="search"><Search size={16}/><input value={query} onChange={event=>setQuery(event.target.value)} placeholder={t.search}/><kbd>⌘ K</kbd></label><span className="top-divider"/><button className="icon-button" aria-label={t.language} onClick={()=>changeLanguage(lang==='en'?'zh-CN':lang==='zh-CN'?'ja':'en')}><Globe size={17}/></button><button className="profile-avatar" aria-label={t.profile}>P</button></div></header>
   <section className="content">
    {page==='home'&&<>
     <div className="hero"><div className="hero-copy"><div className="eyebrow"><span/> {t.personalEyebrow}</div><h1>{t.hero.split('\n').map((row,index)=><React.Fragment key={row}>{index>0&&<br/>}{row}</React.Fragment>)}</h1><p>{t.subtitle}</p><button className="primary" onClick={()=>setPage('favorites')}>{t.browse}<ChevronRight size={16}/></button><div className="hero-note"><span className="avatar mini">PV</span><span>{t.source}</span></div></div><div className="hero-art"><div className="art-orbit orbit-one"/><div className="art-orbit orbit-two"/><div className="art-card back-card"><div className="art-sun"/><div className="art-mountain"/><div className="art-water"/></div><div className="art-card front-card"><div className="art-sun small"/><div className="art-mountain"/><div className="art-water"/><span className="art-play"><Play size={16} fill="currentColor"/></span></div><div className="art-label"><span className="green-dot"/> {t.offline}</div></div><div className="hero-grain"/></div>
     <div className="stats"><div><div className="stat-icon"><Library size={17}/></div><small>{t.all}</small><strong>{videos.length.toString().padStart(2,'0')} <em>{t.videosUnit}</em></strong></div><div><div className="stat-icon"><HardDrive size={17}/></div><small>{t.saved}</small><strong>{videos.filter(video=>video.has_video).length.toString().padStart(2,'0')} <em>{t.available}</em></strong></div><div><div className="stat-icon"><Clock size={17}/></div><small>{t.recent}</small><strong>{videos.length?t.today:'—'} <em>{t.synced}</em></strong></div></div>
     <div className="section-heading"><div><span className="eyebrow">{t.collectionEyebrow}</span><h2>{t.recent}</h2></div><button className="text-button" onClick={()=>setPage('videos')}>{t.all} <ChevronRight size={15}/></button></div>
     <VideoGrid videos={videos.slice(0,3)} open={openVideo} download={startDownload} t={t}/>
    </>}
    {(page==='videos'||page==='favorites')&&<>
     <div className="page-heading"><div><span className="eyebrow">{t.collectionEyebrow}</span><h1>{page==='favorites'?t.favorites:t.all}</h1><p>{t.library}</p></div><button className="outline" onClick={()=>api('/sync',{method:'POST'}).then(refresh)}><RefreshCw size={15}/>{t.sync}</button></div>
     <VideoGrid videos={videos} open={openVideo} download={startDownload} t={t}/>
    </>}
    {page==='detail'&&selected&&<>
     <button className="back-button" onClick={()=>setPage('videos')}><ArrowLeft size={15}/>{t.back}</button>
     <div className="detail-layout"><div className="detail-cover">{selected.has_video?<video controls playsInline src={'/api/media/'+selected.id}/>:<div className="cover-landscape"><span className="sun"/><div className="mountains"/><div className="lake"/></div>}{!selected.has_video&&<div className="cover-overlay"><button className="play-large" onClick={()=>startDownload(selected)} aria-label={t.download}><Download/></button><span>{t.noVideo}</span></div>}</div><div className="detail-info"><span className="eyebrow">{t.favoriteEyebrow}</span><h1>{selected.title}</h1><p>{selected.creator} · {formatDuration(selected.duration)}</p><div className="detail-rule"/><button className="primary" onClick={()=>startDownload(selected)}><Download size={16}/>{t.download}</button><button className="outline" onClick={()=>startDownload(selected,'audio')}><Download size={16}/>{t.audio}</button><div className="detail-meta"><span>{t.status}</span><b>{selected.has_video?t.offline:t.noVideo}</b></div></div></div>
    </>}
    {page==='queue'&&<>
     <div className="page-heading"><div><span className="eyebrow">{t.transferEyebrow}</span><h1>{t.queue}</h1><p>{t.source}</p></div></div>
     <div className="queue-list">{tasks.length===0?<div className="empty-state"><div className="empty-icon"><Download/></div><h3>{t.empty}</h3><p>{t.library}</p><button className="outline" onClick={()=>setPage('videos')}>{t.all}</button></div>:tasks.map(task=><div className="task" key={task.id}><div className="task-icon"><Download size={17}/></div><div className="task-main"><b>{videos.find(video=>video.id===task.video_id)?.title||task.video_id}</b><small>{t[statusKey[task.status]||'queued']} · {task.kind==='audio'?t.audioKind:t.videoKind}</small><div className="progress"><span style={{width:task.progress+'%'}}/></div></div><strong>{task.progress}%</strong><button className="icon-button" aria-label={task.status==='running'?'Pause':'Resume'} onClick={async()=>{await api('/task/'+task.id+'/'+(task.status==='running'?'pause':'resume'),{method:'POST'});await refresh()}}>{task.status==='running'?<Pause size={15}/>:<Play size={15}/>}</button></div>)}</div>
    </>}
    {page==='settings'&&<>
     <div className="page-heading"><div><span className="eyebrow">{t.makeItYoursEyebrow}</span><h1>{t.settings}</h1><p>{t.preferences}</p></div></div>
     <div className="settings-card"><div className="settings-title"><div className="setting-symbol"><SettingsIcon size={18}/></div><div><h3>{t.preferences}</h3><p>{t.escape}</p></div></div><div className="setting-row"><div><b>{t.language}</b><small>{t.languageHint}</small></div><select value={lang} onChange={event=>changeLanguage(event.target.value)}><option value="en">English</option><option value="zh-CN">简体中文</option><option value="ja">日本語</option></select></div><div className="setting-row"><div><b>{t.theme}</b><small>{t.themeHint}</small></div><div className="theme-options">{[['system',Monitor],['light',Sun],['dark',Moon]].map(([value,Icon]:any)=><button className={settings.theme===value?'chosen':''} onClick={()=>chooseTheme(value)} key={value}><Icon size={15}/>{t[value as keyof Copy]}</button>)}</div></div><div className="setting-row"><div><b>{t.cache}</b><small>{t.cacheHint}</small></div><input value={settings.cache_dir} onChange={event=>setSettings({...settings,cache_dir:event.target.value})}/></div><div className="setting-row"><div><b>{t.folder}</b><small>{t.folderHint}</small></div><input value={settings.download_dir} onChange={event=>setSettings({...settings,download_dir:event.target.value})}/></div><div className="setting-row"><div><b>{t.frequency}</b><small>{t.frequencyHint}</small></div><input type="number" min="1" value={settings.sync_minutes} onChange={event=>setSettings({...settings,sync_minutes:+event.target.value})}/></div><div className="setting-row"><div><b>{t.automatic}</b><small>{t.automaticHint}</small></div><button className={'toggle '+(settings.auto_sync?'on':'')} aria-label={t.automatic} aria-pressed={settings.auto_sync} onClick={()=>setSettings({...settings,auto_sync:!settings.auto_sync})}><span/></button></div><div className="setting-row"><div><b>{t.proxy}</b><small>{t.proxyHint}</small></div><input placeholder="http://127.0.0.1:7890" value={settings.proxy} onChange={event=>setSettings({...settings,proxy:event.target.value})}/></div><div className="save-row"><span>{saved&&<><Check size={15}/>{t.savedmsg}</>}</span><button className="primary" onClick={saveSettings}>{t.save}<ChevronRight size={15}/></button></div></div>
    </>}
   </section>
  </main>
 </div>
}
createRoot(document.getElementById('root')!).render(<App/>);
