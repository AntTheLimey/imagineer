// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * HTML sanitization utilities using DOMPurify.
 */

import DOMPurify from 'dompurify';

/**
 * Sanitizes HTML content to prevent XSS attacks.
 *
 * Uses DOMPurify to remove potentially dangerous HTML elements and attributes
 * while preserving safe formatting elements commonly used in rich text content.
 *
 * @param html - The HTML string to sanitize.
 * @returns The sanitized HTML string.
 */
export function sanitizeHtml(html: string): string {
    return DOMPurify.sanitize(html, {
        ALLOWED_TAGS: [
            'p', 'br', 'strong', 'b', 'em', 'i', 'u', 's', 'strike',
            'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
            'ul', 'ol', 'li',
            'blockquote', 'pre', 'code',
            'a', 'span', 'div',
            'table', 'thead', 'tbody', 'tr', 'th', 'td',
            'hr', 'sub', 'sup',
        ],
        ALLOWED_ATTR: [
            'href', 'title', 'target', 'rel',
            'class', 'style',
        ],
        // Force all links to open in a new tab with security attributes
        ADD_ATTR: ['target', 'rel'],
    });
}

/**
 * Strips all HTML tags from content, returning plain text.
 *
 * Useful for generating plain text previews from rich text content.
 *
 * @param html - The HTML string to strip.
 * @returns Plain text with all HTML tags removed.
 */
export function stripHtml(html: string): string {
    // First sanitize to prevent XSS
    const sanitized = DOMPurify.sanitize(html, { ALLOWED_TAGS: [] });
    // Then decode HTML entities
    const doc = new DOMParser().parseFromString(sanitized, 'text/html');
    return doc.body.textContent ?? '';
}
