/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsSafeDisplayStringConstraint } from './is-safe-string.validator'

describe('IsSafeDisplayStringConstraint', () => {
  const validator = new IsSafeDisplayStringConstraint()

  describe('valid inputs', () => {
    const validStrings = [
      ['simple name', 'My Organization'],
      ['name with apostrophe', "John O'Brien-Smith"],
      ['name with dot', 'Acme Inc.'],
      ['name with underscore', 'my_sandbox_123'],
      ['hyphenated identifier', 'us-east-1'],
      ['snapshot name', 'ubuntu-4vcpu-8ram-100gb'],
      ['docker image name', 'ubuntu:22.04'],
      ['docker image with registry', 'ghcr.io/org/image:latest'],
      ['CIDR notation', '192.168.1.0/16,10.0.0.0/24'],
      ['email address', 'user@example.com'],
      ['unicode name', 'Ünternehmen'],
      ['CJK characters', '\u65E5\u672C\u8A9E\u30C6\u30B9\u30C8'],
      ['cyrillic name', '\u041E\u0440\u0433\u0430\u043D\u0438\u0437\u0430\u0446\u0438\u044F'],
      ['empty string', ''],
      ['single character', 'A'],
      ['numbers only', '12345'],
      ['path-like string', '/usr/local/bin'],
      ['version string', 'v1.2.3-beta.1'],
      ['entrypoint command', 'sleep infinity'],
      ['hash value', 'abc123def456789'],
      ['role ID', 'developer'],
      ['region ID', 'eu-west-1'],
      ['api key format', 'sk-1234567890abcdef'],
      ['string with newline', 'line1\nline2'],
      ['string with tab', 'col1\tcol2'],
      ['string with carriage return', 'line1\r\nline2'],
      ['mathematical expression', '3 + 5 = 8'],
      ['special chars allowed', "test._'-value"],
      ['parentheses and brackets', 'value (note) [extra]'],
      ['curly braces', '{ key: value }'],
      ['pipe character', 'a | b'],
      ['at sign', 'user@domain'],
      ['hash sign', '#channel'],
      ['dollar sign', '$variable'],
      ['percent sign', '50%'],
      ['ampersand in text', 'Tom & Jerry'],
      ['equals sign', 'key=value'],
      ['plus sign', 'a+b'],
      ['tilde', '~user'],
      ['backtick', '`code`'],
      ['exclamation mark', 'Hello!'],
      ['question mark', 'What?'],
      ['semicolon', 'a; b'],
      ['colon', 'key: value'],
      ['comma', 'a, b, c'],
      ['less-than not followed by letter', '3 < 5'],
      ['greater-than alone', '5 > 3'],
      ['angle brackets with space', '< not a tag >'],
    ]

    it.each(validStrings)('should accept %s: %s', (_description, value) => {
      expect(validator.validate(value)).toBe(true)
    })
  })

  describe('HTML tag rejection', () => {
    const htmlStrings = [
      ['script tag', '<script>alert(1)</script>'],
      ['img tag', '<img src=x onerror=alert(1)>'],
      ['anchor tag', '<a href="evil">click</a>'],
      ['div tag', '<div>content</div>'],
      ['self-closing br', '<br/>'],
      ['self-closing img', '<img />'],
      ['bold tag', '<b>bold</b>'],
      ['closing tag only', '</div>'],
      ['tag embedded in text', 'Hello <b>world</b>'],
      ['tag at start', '<span>text'],
      ['tag at end', 'text<span>'],
      ['tag with attributes', '<input type="text" value="test">'],
      ['tag with single quotes', "<img src='x'>"],
      ['SVG tag', '<svg onload=alert(1)>'],
      ['iframe tag', '<iframe src="evil.com">'],
      ['link tag', '<link rel="stylesheet" href="evil.css">'],
      ['style tag', '<style>body{display:none}</style>'],
      ['object tag', '<object data="evil.swf">'],
      ['embed tag', '<embed src="evil.swf">'],
      ['form tag', '<form action="evil.com">'],
      ['unclosed img tag (XSS bypass)', '<img src=x onerror=alert(1)'],
      ['unclosed script tag', '<script type="text/javascript"'],
      ['unclosed tag at end of string', 'hello <b'],
    ]

    it.each(htmlStrings)('should reject %s: %s', (_description, value) => {
      expect(validator.validate(value)).toBe(false)
    })
  })

  describe('URL rejection', () => {
    const urlStrings = [
      ['https URL', 'https://evil.com'],
      ['http URL', 'http://phishing.site'],
      ['ftp URL', 'ftp://files.example.com'],
      ['https with path', 'https://evil.com/path/to/page'],
      ['http with port', 'http://evil.com:8080'],
      ['https with query', 'https://evil.com?q=test'],
      ['www prefix', 'www.evil.com'],
      ['URL embedded in text', 'Visit https://evil.com for details'],
      ['URL at end of text', 'Check http://phish.me'],
      ['URL at start of text', 'https://evil.com is bad'],
      ['www in sentence', 'Go to www.evil.com now'],
      ['mixed case HTTP', 'HTTP://EVIL.COM'],
      ['mixed case Https', 'Https://Evil.com'],
      ['FTP uppercase', 'FTP://files.server.com'],
      ['URL with fragment', 'https://evil.com#section'],
      ['URL with auth', 'https://user:pass@evil.com'],
      ['URL with encoded chars', 'https://evil.com/%20path'],
      ['localhost URL', 'http://localhost:3000'],
      ['IP URL', 'https://192.168.1.1'],
      ['WWW uppercase', 'WWW.evil.com'],
      ['ssh URL', 'ssh://evil.com'],
      ['sftp URL', 'sftp://files.evil.com'],
      ['ftps URL', 'ftps://files.evil.com'],
      ['websocket URL', 'ws://evil.com/socket'],
      ['secure websocket URL', 'wss://evil.com/socket'],
      ['file URL', 'file:///etc/passwd'],
      ['ldap URL', 'ldap://evil.com/dc=example'],
      ['ldaps URL', 'ldaps://evil.com/dc=example'],
      ['javascript pseudo-URL', 'javascript:alert(1)'],
      ['data URI', 'data:text/html,<script>alert(1)</script>'],
      ['mailto link', 'mailto:evil@example.com'],
      ['tel link', 'tel:+1234567890'],
      ['javascript in text', 'click javascript:void(0)'],
      ['data URI in text', 'see data:image/png;base64,abc'],
      ['SSH uppercase', 'SSH://evil.com'],
    ]

    it.each(urlStrings)('should reject %s: %s', (_description, value) => {
      expect(validator.validate(value)).toBe(false)
    })
  })

  describe('control character rejection', () => {
    const controlCharStrings = [
      ['null byte', '\x00test'],
      ['SOH', 'test\x01'],
      ['STX', '\x02test'],
      ['ETX', 'test\x03'],
      ['BEL', 'test\x07'],
      ['backspace', 'test\x08'],
      ['vertical tab', 'test\x0B'],
      ['form feed', 'test\x0C'],
      ['SO', 'test\x0E'],
      ['SI', 'test\x0F'],
      ['DLE', 'test\x10'],
      ['ESC', 'test\x1B'],
      ['US', 'test\x1F'],
      ['DEL', 'test\x7F'],
      ['null in middle', 'te\x00st'],
      ['multiple control chars', '\x01\x02\x03'],
    ]

    it.each(controlCharStrings)('should reject %s', (_description, value) => {
      expect(validator.validate(value)).toBe(false)
    })
  })

  describe('allowed whitespace characters', () => {
    it('should accept tab (\\x09)', () => {
      expect(validator.validate('col1\tcol2')).toBe(true)
    })

    it('should accept line feed (\\x0A)', () => {
      expect(validator.validate('line1\nline2')).toBe(true)
    })

    it('should accept carriage return (\\x0D)', () => {
      expect(validator.validate('line1\rline2')).toBe(true)
    })

    it('should accept CRLF', () => {
      expect(validator.validate('line1\r\nline2')).toBe(true)
    })
  })

  describe('nullish values (safe — allows missing optional fields)', () => {
    it('should accept undefined', () => {
      expect(validator.validate(undefined)).toBe(true)
    })

    it('should accept null', () => {
      expect(validator.validate(null)).toBe(true)
    })
  })

  describe('non-string type rejection', () => {
    it('should reject number', () => {
      expect(validator.validate(123)).toBe(false)
    })

    it('should reject object', () => {
      expect(validator.validate({ key: 'value' })).toBe(false)
    })

    it('should reject array', () => {
      expect(validator.validate(['a', 'b'])).toBe(false)
    })

    it('should reject boolean', () => {
      expect(validator.validate(true)).toBe(false)
    })
  })

  describe('combined attack patterns', () => {
    const combinedAttacks = [
      ['URL + HTML', 'https://evil.com<script>alert(1)</script>'],
      ['HTML + control char', '<img src=x>\x00'],
      ['URL + control char', 'https://evil.com\x00'],
      ['URL in HTML attribute', '<a href="https://evil.com">click</a>'],
      ['script with URL', '<script src="https://evil.com/xss.js"></script>'],
      ['encoded XSS attempt', '<img src=x onerror="fetch(\'https://evil.com\')">'],
    ]

    it.each(combinedAttacks)('should reject %s: %s', (_description, value) => {
      expect(validator.validate(value)).toBe(false)
    })
  })

  describe('defaultMessage', () => {
    it('should include property name when available', () => {
      const args = { property: 'name' } as import('class-validator').ValidationArguments
      expect(validator.defaultMessage(args)).toBe('name must not contain HTML tags, URLs, or control characters')
    })

    it('should fall back to Value when no args provided', () => {
      expect(validator.defaultMessage()).toBe('Value must not contain HTML tags, URLs, or control characters')
    })
  })
})
