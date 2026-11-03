// Minimal client-side PDF writer for Seed Poet. Builds a one-page A4 PDF with the
// built-in Helvetica fonts, entirely in the browser: the poem never leaves the page.
// No third-party code (strict CSP), ASCII text only, so byte offsets equal string length.
(function () {
  'use strict';

  function esc(t) {
    return String(t).replace(/[^\x20-\x7e]/g, '?').replace(/([\\()])/g, '\\$1');
  }

  function text(font, size, x, y, t) {
    return 'BT /' + font + ' ' + size + ' Tf 1 0 0 1 ' + x + ' ' + y + ' Tm (' + esc(t) + ') Tj ET\n';
  }

  // lines: array of strings, each "verb adj noun".
  function buildPoemPDF(lines) {
    var words = lines.join(' ').split(/\s+/).filter(Boolean);
    var n = words.length;
    var c = '';
    var y = 780;
    c += '0.09 0.4 0.2 rg\n';
    c += text('F2', 26, 60, y, 'Seed Poet');
    c += '0.3 0.3 0.3 rg\n';
    y -= 22;
    c += text('F1', 12, 60, y, n + '-word recovery poem (BIP39)');
    c += '0 0 0 rg\n';

    y -= 44;
    for (var i = 0; i < lines.length; i++) {
      c += text('F1', 17, 72, y, lines[i]);
      y -= (i % 2 === 1) ? 34 : 24;
    }

    y -= 6;
    c += text('F2', 12, 60, y, 'Words in order');
    y -= 22;
    var rows = Math.ceil(n / 3);
    for (var k = 0; k < n; k++) {
      var col = Math.floor(k / rows), row = k % rows;
      var label = (k + 1) + '. ' + words[k];
      c += text('F1', 11, 72 + col * 160, y - row * 17, label);
    }

    var notes = [
      'Anyone who has these words controls the wallet. Keep this paper somewhere safe and private.',
      'Print it, then delete this PDF file and empty the trash. Do not keep it in cloud storage or email.',
      'Seed Poet never stored this poem. Strength: about ' + (n === 24 ? '176' : '88') +
        ' bits (a fully random ' + n + '-word BIP39 phrase has ' + (n === 24 ? '256' : '128') + ').',
      'seedpoet.com - developed by OxiSoft'
    ];
    c += '0.3 0.3 0.3 rg\n';
    var ny = 110;
    for (var m = 0; m < notes.length; m++) {
      c += text('F1', 9, 60, ny, notes[m]);
      ny -= 14;
    }

    var objs = [
      '<< /Type /Catalog /Pages 2 0 R >>',
      '<< /Type /Pages /Kids [3 0 R] /Count 1 >>',
      '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] ' +
        '/Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>',
      '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>',
      '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>',
      '<< /Length ' + c.length + ' >>\nstream\n' + c + 'endstream'
    ];
    var out = '%PDF-1.4\n';
    var offsets = [];
    for (var o = 0; o < objs.length; o++) {
      offsets.push(out.length);
      out += (o + 1) + ' 0 obj\n' + objs[o] + '\nendobj\n';
    }
    var xref = out.length;
    out += 'xref\n0 ' + (objs.length + 1) + '\n0000000000 65535 f \n';
    for (var q = 0; q < offsets.length; q++) {
      out += ('0000000000' + offsets[q]).slice(-10) + ' 00000 n \n';
    }
    out += 'trailer\n<< /Size ' + (objs.length + 1) + ' /Root 1 0 R >>\nstartxref\n' + xref + '\n%%EOF\n';
    return new Blob([out], { type: 'application/pdf' });
  }

  window.SeedPoetPDF = { build: buildPoemPDF };
})();
