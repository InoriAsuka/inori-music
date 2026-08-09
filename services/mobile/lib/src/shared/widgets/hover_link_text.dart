import 'package:flutter/material.dart';

/// A [Text] that behaves like a hyperlink only while the pointer is hovering
/// it, and only when [onTap] is non-null — an underline appears and the
/// cursor becomes a pointing hand; at rest it is pixel-identical to plain
/// text.
///
/// This is deliberately *not* the usual always-blue, always-underlined link
/// treatment: this app's whole visual direction is a cover-driven colour
/// field with a single sakura-pink accent (see the
/// visual-direction-cover-driven memory), and a permanently blue/underlined
/// track title would clash with that everywhere it appeared rather than
/// reading as an incidental affordance. Hover is also the only signal
/// available that costs nothing when it isn't there: a touch user gets
/// exactly the same look this text always had (MouseRegion never fires for
/// touch input — the same reasoning local_library_screen.dart's own
/// `_hovering` field already relies on), so tapping remains the only way to
/// discover the link there, same as every tappable list row elsewhere in
/// this app.
///
/// [onTap] is nullable on purpose rather than requiring the caller to branch
/// between this widget and a plain [Text]: a local track (guest mode) has no
/// album/artist id to link to, and the v5.30.7 field report was explicit
/// that showing a link-looking control that does nothing when tapped is
/// worse than not showing one — passing `onTap: null` here renders inert,
/// unstyled plain text with no [MouseRegion]/[GestureDetector] wrapping it at
/// all, so there is no hover/cursor affordance promising an action that
/// isn't there.
class HoverLinkText extends StatefulWidget {
  const HoverLinkText({
    super.key,
    required this.text,
    required this.style,
    this.onTap,
    this.textAlign,
    this.maxLines,
    this.overflow,
  });

  final String text;
  final TextStyle style;
  final VoidCallback? onTap;
  final TextAlign? textAlign;
  final int? maxLines;
  final TextOverflow? overflow;

  @override
  State<HoverLinkText> createState() => _HoverLinkTextState();
}

class _HoverLinkTextState extends State<HoverLinkText> {
  bool _hovering = false;

  @override
  Widget build(BuildContext context) {
    final linkable = widget.onTap != null;
    final text = Text(
      widget.text,
      style: linkable && _hovering
          ? widget.style.copyWith(decoration: TextDecoration.underline)
          : widget.style,
      textAlign: widget.textAlign,
      maxLines: widget.maxLines,
      overflow: widget.overflow,
    );
    if (!linkable) return text;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: widget.onTap,
        child: text,
      ),
    );
  }
}
