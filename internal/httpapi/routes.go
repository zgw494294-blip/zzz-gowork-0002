package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *Handler) serveWeb(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// serve index.html from embedded or file
	// For simplicity, we read from internal/web/index.html
	// In real deployment, path should be relative to executable
	wd, err := os.Getwd()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path := filepath.Join(wd, "internal", "web", "index.html")
	if _, err := os.Stat(path); err != nil {
		// fallback to embedded string
		http.ServeContent(w, r, "index.html", timeNow(), strings.NewReader(indexHTML))
		return
	}
	http.ServeFile(w, r, path)
}

func timeNow() time.Time {
	return time.Now()
}

const indexHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8"><title>水质采样链路</title>
<style>body{font-family:sans-serif;margin:40px}input,button{margin:5px}section{margin-bottom:20px}</style>
</head>
<body>
<h1>水质采样链路服务</h1>

<section>
<h2>创建采样点</h2>
<input id="ptName" placeholder="名称"><input id="ptLoc" placeholder="位置"><button onclick="createPoint()">创建</button>
</section>

<section>
<h2>创建批次</h2>
<select id="pointSelect"></select><input id="batchNo" placeholder="批次号"><button onclick="createBatch()">创建</button>
</section>

<section>
<h2>批次操作</h2>
<select id="batchSelect"></select>
<br>
<input id="bottleCode" placeholder="样本瓶代码"><button onclick="addBottle()">添加样本瓶</button>
<button onclick="completeSampling()">完成采样</button>
<br>
<input id="handoverBottleId" placeholder="瓶ID"><input id="from" placeholder="从"><input id="to" placeholder="至"><input id="note" placeholder="备注"><button onclick="handover()">交接</button>
<br>
<input id="conclusion" placeholder="结论"><button onclick="confirmConclusion()">确认结论</button>
</section>

<section>
<h2>汇总</h2>
<button onclick="loadSummary()">刷新汇总</button>
<pre id="summary"></pre>
</section>

<script>
async function api(url, method, body) {
  const opt = {method: method, headers:{'Content-Type':'application/json'}};
  if(body) opt.body = JSON.stringify(body);
  const resp = await fetch(url, opt);
  if(!resp.ok) { const err = await resp.json(); throw new Error(err.error); }
  return resp.json();
}
async function refreshPoints() {
  const pts = await api('/api/points','GET');
  const sel = document.getElementById('pointSelect');
	  sel.innerHTML = pts.map(p=>'<option value="' + p.id + '">' + p.name + '</option>').join('');
}
async function refreshBatches() {
  const bs = await api('/api/batches','GET');
  const sel = document.getElementById('batchSelect');
	  sel.innerHTML = bs.map(b=>'<option value="' + b.id + '">' + b.batch_no + '</option>').join('');
}
async function createPoint() {
  try { await api('/api/points','POST',{name:ptName.value,location:ptLoc.value}); refreshPoints(); } catch(e){ alert(e.message); }
}
async function createBatch() {
  try { await api('/api/batches','POST',{point_id:pointSelect.value,batch_no:batchNo.value}); refreshBatches(); } catch(e){ alert(e.message); }
}
async function addBottle() {
  try { await api('/api/batches/'+batchSelect.value+'/bottles','POST',{code:bottleCode.value}); } catch(e){ alert(e.message); }
}
async function completeSampling() {
  try { await api('/api/batches/'+batchSelect.value+'/sampling','POST',{}); } catch(e){ alert(e.message); }
}
async function handover() {
  try { await api('/api/batches/'+batchSelect.value+'/handover','POST',{bottle_id:handoverBottleId.value,from:from.value,to:to.value,note:note.value}); } catch(e){ alert(e.message); }
}
async function confirmConclusion() {
  try { await api('/api/batches/'+batchSelect.value+'/conclusion','POST',{conclusion:conclusion.value}); } catch(e){ alert(e.message); }
}
async function loadSummary() {
  try { const s = await api('/api/samples/summary','GET'); document.getElementById('summary').textContent = JSON.stringify(s, null, 2); } catch(e){ alert(e.message); }
}
refreshPoints(); refreshBatches();
</script>
</body>
</html>`
