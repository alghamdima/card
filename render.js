/* Shared card renderer — used by index.html (preview + export) and admin.html
   so the preview is always identical to the downloaded PNG. */
(function (root) {

  var CANVAS_W = 1080, CANVAS_H = 1350;
  var SEP = ' \u2014 ';          // heading / body separator
  var SECTION_GAP = 0.7;         // gap between messages, as a share of font size

  function isArabic(t) { return /[\u0600-\u06FF]/.test(t || ''); }

  function fontFor(text, size, weight) {
    var ar = isArabic(text);
    var fam = (weight === 'bold') ? (ar ? 'LumaSemiBold' : 'KarbonBold')
                                  : (ar ? 'LumaRegular'  : 'KarbonRegular');
    return size + 'px "' + fam + '"';
  }

  /* Wrap body text, reserving room on the first line for an inline heading. */
  function wrapWithPrefix(ctx, text, maxW, firstLineIndent) {
    var words = String(text).split(/\s+/).filter(Boolean);
    var lines = [], cur = '', limit = maxW - firstLineIndent;
    for (var i = 0; i < words.length; i++) {
      var w = words[i];
      var test = cur ? cur + ' ' + w : w;
      if (ctx.measureText(test).width <= limit) { cur = test; continue; }
      if (cur) { lines.push(cur); cur = ''; limit = maxW; }
      if (ctx.measureText(w).width > limit) {          // single word too long
        var piece = '';
        for (var j = 0; j < w.length; j++) {
          var t2 = piece + w[j];
          if (ctx.measureText(t2).width > limit && piece) { lines.push(piece); piece = w[j]; limit = maxW; }
          else piece = t2;
        }
        cur = piece;
      } else cur = w;
    }
    if (cur) lines.push(cur);
    return lines.length ? lines : [''];
  }

  function wrapPlain(ctx, text, maxW) { return wrapWithPrefix(ctx, text, maxW, 0); }

  /* Lay out N sections of "Heading — body" at one shared font size. */
  function fitSections(ctx, sections, box) {
    var maxSize = Number(box.size) || 40;
    var minSize = Math.max(9, Math.round(maxSize * 0.32));
    var boxH = box.bottom - box.top;

    for (var size = maxSize; size >= minSize; size--) {
      var lh = size * 1.32, gap = size * SECTION_GAP;
      var plan = [], total = 0, ok = true;

      for (var i = 0; i < sections.length; i++) {
        var head = (sections[i].heading || '').trim();
        var body = (sections[i].text || '').trim();
        if (!head && !body) continue;

        var indent = 0;
        if (head) {
          ctx.font = fontFor(head, size, 'bold');
          indent = ctx.measureText(head + SEP).width;
        }
        ctx.font = fontFor(body, size, 'regular');
        var lines = body ? wrapWithPrefix(ctx, body, box.maxW, indent) : [''];

        // a heading that alone exceeds the box can't be laid out at this size
        if (indent > box.maxW * 0.9) { ok = false; break; }

        plan.push({ head: head, indent: indent, lines: lines, size: size });
        total += lines.length * lh;
        if (plan.length > 1) total += gap;
      }

      if (!plan.length) return null;
      if ((ok && total <= boxH) || size === minSize) {
        return { plan: plan, total: total, lh: lh, gap: gap, size: size };
      }
    }
    return null;
  }

  function drawSections(ctx, sections, box) {
    var fit = fitSections(ctx, sections, box);
    if (!fit) return;
    var headColor = box.headColor || box.color || '#FFFFFF';
    var bodyColor = box.bodyColor || box.color || '#FFFFFF';
    ctx.textBaseline = 'middle';
    var y = box.top + (box.bottom - box.top - fit.total) / 2;

    fit.plan.forEach(function (sec, si) {
      sec.lines.forEach(function (line, li) {
        var cy = y + fit.lh / 2;
        if (li === 0 && sec.head) {
          ctx.font = fontFor(line, fit.size, 'regular');
          var restW = ctx.measureText(line).width;
          var startX = box.centerX - (sec.indent + restW) / 2;
          ctx.textAlign = 'left';
          ctx.font = fontFor(sec.head, fit.size, 'bold');
          ctx.fillStyle = headColor;
          ctx.fillText(sec.head + SEP, startX, cy);
          ctx.font = fontFor(line, fit.size, 'regular');
          ctx.fillStyle = bodyColor;
          ctx.fillText(line, startX + sec.indent, cy);
          ctx.textAlign = 'center';
        } else {
          ctx.textAlign = 'center';
          ctx.font = fontFor(line, fit.size, 'regular');
          ctx.fillStyle = bodyColor;
          ctx.fillText(line, box.centerX, cy);
        }
        y += fit.lh;
      });
      if (si < fit.plan.length - 1) y += fit.gap;
    });
  }

  /* Simple single-string box (To / From). */
  function drawBox(ctx, text, box) {
    if (!text || !box) return;
    var maxSize = Number(box.size) || 40;
    var minSize = Math.max(9, Math.round(maxSize * 0.35));
    var weight = box.weight === 'regular' ? 'regular' : 'bold';
    var boxH = box.bottom - box.top, chosen = null;
    for (var size = maxSize; size >= minSize; size--) {
      ctx.font = fontFor(text, size, weight);
      var lines = wrapPlain(ctx, text, box.maxW);
      var lh = size * 1.3, total = lines.length * lh;
      var widest = 0;
      lines.forEach(function (l) { widest = Math.max(widest, ctx.measureText(l).width); });
      if ((total <= boxH && widest <= box.maxW) || size === minSize) { chosen = { lines: lines, lh: lh, total: total, size: size }; break; }
    }
    if (!chosen) return;
    ctx.font = fontFor(text, chosen.size, weight);
    ctx.fillStyle = box.color || '#FFFFFF';
    ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
    var y = box.top + (boxH - chosen.total) / 2;
    chosen.lines.forEach(function (l) { ctx.fillText(l, box.centerX, y + chosen.lh / 2); y += chosen.lh; });
  }

  /* data = { to, from, messages:[{heading,text},...] }
     also accepts the older { heading, message } shape. */
  function render(canvas, img, boxes, data) {
    var ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);
    if (img && img.complete && img.naturalWidth) ctx.drawImage(img, 0, 0, CANVAS_W, CANVAS_H);
    if (!boxes) return;

    drawBox(ctx, (data.to || '').trim(), boxes.to);

    var secs = data.messages;
    if (!secs && (data.heading || data.message)) secs = [{ heading: data.heading, text: data.message }];
    secs = (secs || []).filter(function (s) { return s && ((s.heading || '').trim() || (s.text || '').trim()); });
    if (secs.length && boxes.message) drawSections(ctx, secs, boxes.message);

    var from = (data.from || '').trim();
    if (from) drawBox(ctx, from, boxes.from);
  }

  function defaultBoxes() {
    return {
      to:      { top: 300, bottom: 400,  centerX: 540, maxW: 700, size: 52, color: '#FFFFFF', weight: 'bold' },
      message: { top: 470, bottom: 800,  centerX: 540, maxW: 760, size: 40,
                 headColor: '#F0A868', bodyColor: '#FFFFFF' },
      from:    { top: 900, bottom: 1000, centerX: 540, maxW: 700, size: 44, color: '#FFFFFF', weight: 'bold' }
    };
  }

  root.CardRender = {
    W: CANVAS_W, H: CANVAS_H,
    isArabic: isArabic, render: render, drawBox: drawBox, defaultBoxes: defaultBoxes
  };

})(window);
