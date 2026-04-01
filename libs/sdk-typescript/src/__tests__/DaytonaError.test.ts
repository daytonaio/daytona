import { DaytonaError, DaytonaNotFoundError, DaytonaRateLimitError, DaytonaTimeoutError } from '../errors/DaytonaError'

describe('Daytona errors', () => {
  it('constructs DaytonaError with properties', () => {
    const err = new DaytonaError('boom', 500)
    expect(err).toBeInstanceOf(Error)
    expect(err.name).toBe('DaytonaError')
    expect(err.message).toBe('boom')
    expect(err.statusCode).toBe(500)
  })

  test.each([
    [DaytonaNotFoundError, 'DaytonaNotFoundError'],
    [DaytonaRateLimitError, 'DaytonaRateLimitError'],
    [DaytonaTimeoutError, 'DaytonaTimeoutError'],
  ])('constructs %s', (ErrCtor, expectedName) => {
    const err = new ErrCtor('x', 404)
    expect(err).toBeInstanceOf(DaytonaError)
    expect(err.name).toBe(expectedName)
    expect(err.statusCode).toBe(404)
  })
})
