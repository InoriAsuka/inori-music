import 'package:flutter/material.dart';

import 'package:inori_music/src/shared/theme/sakura_dusk.dart';

/// The Inori Music brand mark: a single notched sakura petal, not a generic
/// music-note glyph. The notch at the outer tip is what reads as
/// specifically "sakura" rather than a generic leaf/teardrop; the two
/// outer curves are deliberately asymmetric (not a mirrored clip-art shape).
///
/// Used on the splash screen, the login screen, and the desktop sidebar so
/// the app has one consistent identity mark instead of a stock icon.
class InoriMark extends StatelessWidget {
  const InoriMark({super.key, this.size = 72});

  final double size;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: size,
      height: size,
      child: CustomPaint(painter: _PetalPainter()),
    );
  }
}

class _PetalPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final scaleX = size.width / 100;
    final scaleY = size.height / 100;

    final petal = Path()
      ..moveTo(50 * scaleX, 93 * scaleY)
      // Left outer bulge, base up to the left shoulder of the notch.
      ..quadraticBezierTo(9 * scaleX, 58 * scaleY, 35 * scaleX, 9 * scaleY)
      // The notch: a shallow inward dip between the two shoulders — this is
      // what makes the silhouette read as a sakura petal rather than a
      // generic leaf/teardrop.
      ..quadraticBezierTo(49 * scaleX, 23 * scaleY, 50 * scaleX, 19 * scaleY)
      ..quadraticBezierTo(52 * scaleX, 24 * scaleY, 66 * scaleX, 7 * scaleY)
      // Right outer bulge back down to the base — intentionally not a
      // mirror of the left side (slightly wider sweep) to avoid a
      // perfectly-symmetric clip-art look.
      ..quadraticBezierTo(95 * scaleX, 52 * scaleY, 50 * scaleX, 93 * scaleY)
      ..close();

    final gradient = LinearGradient(
      begin: Alignment.topRight,
      end: Alignment.bottomLeft,
      colors: [SakuraDuskColors.sakuraPinkLight, SakuraDuskColors.sakuraPinkDark],
    );
    final fillPaint = Paint()
      ..shader = gradient.createShader(Rect.fromLTWH(0, 0, size.width, size.height));
    canvas.drawPath(petal, fillPaint);

    // A single etched groove line — a quiet nod to "music" (grooves/sound
    // waves) without resorting to a literal note glyph.
    final groove = Path()
      ..moveTo(38 * scaleX, 72 * scaleY)
      ..quadraticBezierTo(50 * scaleX, 50 * scaleY, 60 * scaleX, 34 * scaleY);
    final groovePaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = size.width * 0.035
      ..strokeCap = StrokeCap.round
      ..color = Colors.white.withValues(alpha: 0.55);
    canvas.drawPath(groove, groovePaint);
  }

  @override
  bool shouldRepaint(covariant _PetalPainter oldDelegate) => false;
}
