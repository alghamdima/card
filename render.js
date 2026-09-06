/* Shared card renderer — used by both index.html and admin.html
   so the admin preview is pixel-identical to what employees get. */
(function (root) {

  var CANVAS_W = 1080, CANVAS_H = 1350;

  function isArabic(t) { return /[\u0600-\u06FF]/.test(t || ''); }

  function fontFor(text, size, weight) {
    var ar = isArabic(text);
    var fam;
    if (weight === 'bold') fam = ar ? 'LumaSemiBold' : 'KarbonBold';
    else fam = ar ? 'LumaRegular' : 'KarbonRegular';
    return size + 'px "' + fam + '"';
  }

  // Break text into lines that fit maxW. Falls back to hard character
  // splitting for single words longer than the box.
  function wrapLines(ctx, text, maxW) {
    var paras = String(text).split('\n');
    var out = [];
    paras.forEach(function (para) {
      var words = para.split(/\s+/).filter(Boolean);
      if (!words.length) { out.push(''); return; }
      var cur = '';
      words.forEach(function (w) {
        var test = cur ? cur + ' ' + w : w;
        if (ctx.measureText(test).width <= maxW) { cur = test; return; }
        if (cur) out.push(cur);
        // word alone still too wide -> hard split it
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

  // Shrink until the wrapped block fits the box in both directions.
  function fitBlock(ctx, text, box, weight) {
    var maxSize = Number(box.size) || 40;
    var minSize = Math.max(10, Math.round(maxSize * 0.35));
    var boxH = box.bottom - box.top;
    var size = maxSize;
    while (size >= minSize) {
      ctx.font = fontFor(text, size, weight);
      var lines = wrapLines(ctx, text, box.maxW);
      var lh = size * 1.32;
      var totalH = lines.length * lh;
      var widest = 0;
      lines.forEach(function (l) { widest = Math.max(widest, ctx.measureText(l).width); });
      if (totalH <= boxH && widest <= box.maxW) return { size: size, lines: lines, lh: lh };
      size -= 1;
    }
    ctx.font = fontFor(text, minSize, weight);
    var ls = wrapLines(ctx, text, box.maxW);
    return { size: minSize, lines: ls, lh: minSize * 1.32 };
  }

  function drawBox(ctx, text, box) {
    if (!text || !box) return;
    var weight = box.weight === 'bold' ? 'bold' : 'regular';
    var fit = fitBlock(ctx, text, box, weight);
    ctx.font = fontFor(text, fit.size, weight);
    ctx.fillStyle = box.color || '#FFFFFF';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    var totalH = fit.lines.length * fit.lh;
    var startY = box.top + (box.bottom - box.top - totalH) / 2 + fit.lh / 2;
    fit.lines.forEach(function (line, i) {
      ctx.fillText(line, box.centerX, startY + i * fit.lh);
    });
  }

  // data = { to, message, from }   boxes = { to:{...}, message:{...}, from:{...} }
  function render(canvas, img, boxes, data) {
    var ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);
    if (img && img.complete && img.naturalWidth) ctx.drawImage(img, 0, 0, CANVAS_W, CANVAS_H);
    if (!boxes) return;
    drawBox(ctx, (data.to || '').trim(), boxes.to);
    drawBox(ctx, (data.message || '').trim(), boxes.message);
    // From is omitted entirely when blank — anonymous cards stay clean
    var from = (data.from || '').trim();
    if (from) drawBox(ctx, from, boxes.from);
  }

  function defaultBoxes() {
    return {
      to:      { top: 300, bottom: 400,  centerX: 540, maxW: 700, size: 52, color: '#FFFFFF', weight: 'bold' },
      message: { top: 470, bottom: 800,  centerX: 540, maxW: 760, size: 42, color: '#FFFFFF', weight: 'regular' },
      from:    { top: 900, bottom: 1000, centerX: 540, maxW: 700, size: 44, color: '#FFFFFF', weight: 'bold' }
    };
  }

  root.CardRender = {
    W: CANVAS_W, H: CANVAS_H,
    isArabic: isArabic,
    render: render,
    defaultBoxes: defaultBoxes
  };

})(window);
