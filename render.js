/* Shared card renderer — used by index.html (preview + export) and admin.html
   so the preview is always identical to the downloaded PNG. */
(function (root) {

  var CANVAS_W = 1080, CANVAS_H = 1350;

  function isArabic(t) { return /[\u0600-\u06FF]/.test(t || ''); }

  function fontFor(text, size, weight) {
    var ar = isArabic(text);
    var fam = (weight === 'bold') ? (ar ? 'LumaSemiBold' : 'KarbonBold')
                                  : (ar ? 'LumaRegular'  : 'KarbonRegular');
    return size + 'px "' + fam + '"';
  }

  function wrapLines(ctx, text, maxW) {
    var out = [];
    String(text).split('\n').forEach(function (para) {
      var words = para.split(/\s+/).filter(Boolean);
      if (!words.length) { out.push(''); return; }
      var cur = '';
      words.forEach(function (w) {
        var test = cur ? cur + ' ' + w : w;
        if (ctx.measureText(test).width <= maxW) { cur = test; return; }
        if (cur) out.push(cur);
        if (ctx.measureText(w).width > maxW) {
          var piece = '';
          for (var i = 0; i < w.length; i++) {
            var t2 = piece + w[i];
            if (ctx.measureText(t2).width > maxW && piece) { out.push(piece); piece = w[i]; }
            else piece = t2;
          }
          cur = piece;
        } else cur = w;
      });
      if (cur) out.push(cur);
    });
    return out;
  }

  var HEAD_RATIO = 1.28;
  var GAP_RATIO  = 0.55;

  /* Lay out an optional heading above a body, scaled together to fit the box. */
  function fitBlock(ctx, heading, body, box) {
    var maxSize = Number(box.size) || 40;
    var minSize = Math.max(9, Math.round(maxSize * 0.32));
    var boxH = box.bottom - box.top;

    for (var size = maxSize; size >= minSize; size--) {
      var plan = [], total = 0, widest = 0;

      if (heading) {
        var hs = Math.round(size * HEAD_RATIO);
        ctx.font = fontFor(heading, hs, 'bold');
        var hl = wrapLines(ctx, heading, box.maxW);
        var hlh = hs * 1.24;
        hl.forEach(function (l) { widest = Math.max(widest, ctx.measureText(l).width); });
        plan.push({ lines: hl, size: hs, lh: hlh, weight: 'bold', text: heading, head: true });
        total += hl.length * hlh;
        if (body) total += size * GAP_RATIO;
      }

      if (body) {
        var bw = (box.weight === 'bold') ? 'bold' : 'regular';
        ctx.font = fontFor(body, size, bw);
        var bl = wrapLines(ctx, body, box.maxW);
        var blh = size * 1.32;
        bl.forEach(function (l) { widest = Math.max(widest, ctx.measureText(l).width); });
        plan.push({ lines: bl, size: size, lh: blh, weight: bw, text: body, head: false });
        total += bl.length * blh;
      }

      if (!plan.length) return null;
      if ((total <= boxH && widest <= box.maxW) || size === minSize) {
        return { plan: plan, total: total, gap: size * GAP_RATIO };
      }
    }
    return null;
  }

  function drawBlock(ctx, heading, body, box) {
    if ((!heading && !body) || !box) return;
    var fit = fitBlock(ctx, heading, body, box);
    if (!fit) return;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    var y = box.top + (box.bottom - box.top - fit.total) / 2;
    fit.plan.forEach(function (part) {
      ctx.font = fontFor(part.text, part.size, part.weight);
      ctx.fillStyle = part.head ? (box.color || '#FFFFFF')
                                : (box.bodyColor || box.color || '#FFFFFF');
      part.lines.forEach(function (line) {
        ctx.fillText(line, box.centerX, y + part.lh / 2);
        y += part.lh;
      });
      if (part.head && body) y += fit.gap;
    });
  }

  function drawBox(ctx, text, box) { drawBlock(ctx, '', text, box); }

  /* data = { to, heading, message, from } */
  function render(canvas, img, boxes, data) {
    var ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);
    if (img && img.complete && img.naturalWidth) ctx.drawImage(img, 0, 0, CANVAS_W, CANVAS_H);
    if (!boxes) return;
    drawBlock(ctx, '', (data.to || '').trim(), boxes.to);
    drawBlock(ctx, (data.heading || '').trim(), (data.message || '').trim(), boxes.message);
    var from = (data.from || '').trim();
    if (from) drawBlock(ctx, '', from, boxes.from);
  }

  function defaultBoxes() {
    return {
      to:      { top: 300, bottom: 400,  centerX: 540, maxW: 700, size: 52, color: '#FFFFFF', weight: 'bold' },
      message: { top: 470, bottom: 800,  centerX: 540, maxW: 760, size: 40, color: '#FFFFFF', weight: 'regular' },
      from:    { top: 900, bottom: 1000, centerX: 540, maxW: 700, size: 44, color: '#FFFFFF', weight: 'bold' }
    };
  }

  root.CardRender = {
    W: CANVAS_W, H: CANVAS_H,
    isArabic: isArabic,
    render: render,
    drawBox: drawBox,
    defaultBoxes: defaultBoxes
  };

})(window);
