import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Calculator from './Calculator'
import { ApiError, calculate } from './api'

// Only `calculate` is replaced; importOriginal keeps the real ApiError class so
// the component's `error instanceof ApiError` check behaves as in production.
vi.mock('./api', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./api')>()),
  calculate: vi.fn(),
}))

const mockCalculate = vi.mocked(calculate)

beforeEach(() => {
  mockCalculate.mockReset()
})

const operationSelect = () => screen.getByLabelText('Operation')
const operandA = () => screen.getByLabelText('Operand a')
const operandB = () => screen.getByLabelText('Operand b')
const submitButton = () => screen.getByRole('button')

/** Fills the form by typing, then submits. Empty values are simply not typed. */
async function submit(
  user: ReturnType<typeof userEvent.setup>,
  { operation, a, b }: { operation?: string; a?: string; b?: string },
) {
  if (operation) await user.selectOptions(operationSelect(), operation)
  if (a) await user.type(operandA(), a)
  if (b) await user.type(operandB(), b)
  await user.click(submitButton())
}

describe('Calculator', () => {
  it('renders every operation the backend supports', () => {
    render(<Calculator />)

    // Must match the backend's operation list.
    expect(
      screen.getAllByRole('option').map((option) => option.getAttribute('value')),
    ).toEqual([
      'add',
      'subtract',
      'multiply',
      'divide',
      'power',
      'sqrt',
      'percentage',
    ])
  })

  describe('the second operand', () => {
    it('is disabled for sqrt, which is unary', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await user.selectOptions(operationSelect(), 'sqrt')

      expect(operandB()).toBeDisabled()
    })

    it('is re-enabled when a binary operation is selected again', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await user.selectOptions(operationSelect(), 'sqrt')
      await user.selectOptions(operationSelect(), 'divide')

      expect(operandB()).toBeEnabled()
    })

    // A value next to a disabled field would read as "in use"; b is omitted from
    // the request entirely for a unary operation.
    it('is blanked, not just disabled, when sqrt is selected', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await user.type(operandB(), '42')
      await user.selectOptions(operationSelect(), 'sqrt')

      expect(operandB()).toHaveValue('')
    })

    it('restores the previously typed value when a binary operation is selected again', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await user.type(operandB(), '42')
      await user.selectOptions(operationSelect(), 'sqrt')
      await user.selectOptions(operationSelect(), 'divide')

      expect(operandB()).toHaveValue('42')
    })
  })

  describe('client-side validation', () => {
    // Messages copy the backend's wording so the user cannot tell which layer
    // rejected the input.
    it.each([
      {
        name: 'a is empty',
        fields: { a: '', b: '3' },
        message: 'operand "a" is required',
      },
      {
        name: 'a is only whitespace',
        fields: { a: '   ', b: '3' },
        message: 'operand "a" is required',
      },
      {
        name: 'b is empty for a binary operation',
        fields: { a: '2', b: '' },
        message: 'operand "b" is required for operation "add"',
      },
      {
        name: 'a is not numeric',
        fields: { a: 'abc', b: '3' },
        message: 'operand "a" must be a number',
      },
      {
        name: 'b is not numeric',
        fields: { a: '2', b: 'xyz' },
        message: 'operand "b" must be a number',
      },
      {
        name: 'a is outside the float64 range',
        fields: { a: '1e400', b: '3' },
        message: 'operand "a" must be a finite number',
      },
      {
        name: 'b is outside the float64 range',
        fields: { a: '2', b: '-1e400' },
        message: 'operand "b" must be a finite number',
      },
    ])('reports it when $name, without calling the API', async ({ fields, message }) => {
      const user = userEvent.setup()
      render(<Calculator />)

      await submit(user, fields)

      expect(screen.getByText(message)).toBeInTheDocument()
      // The point of validating here at all: an invalid request never leaves.
      expect(mockCalculate).not.toHaveBeenCalled()
    })

    it('does not require b for sqrt', async () => {
      const user = userEvent.setup()
      mockCalculate.mockResolvedValue({ result: 3 })
      render(<Calculator />)

      await submit(user, { operation: 'sqrt', a: '9' })

      expect(await screen.findByText('Result: 3')).toBeInTheDocument()
      // b is omitted entirely rather than sent as 0.
      expect(mockCalculate).toHaveBeenCalledWith({ operation: 'sqrt', a: 9 })
    })
  })

  describe('submitting', () => {
    it('sends the operands as numbers and renders the result', async () => {
      const user = userEvent.setup()
      mockCalculate.mockResolvedValue({ result: 5 })
      render(<Calculator />)

      await submit(user, { operation: 'add', a: '2', b: '3' })

      expect(mockCalculate).toHaveBeenCalledWith({ operation: 'add', a: 2, b: 3 })
      expect(await screen.findByText('Result: 5')).toBeInTheDocument()
    })

    it('renders a negative and a fractional result unchanged', async () => {
      const user = userEvent.setup()
      mockCalculate.mockResolvedValue({ result: -0.5 })
      render(<Calculator />)

      await submit(user, { operation: 'divide', a: '-1', b: '2' })

      expect(await screen.findByText('Result: -0.5')).toBeInTheDocument()
    })

    it('shows the backend error message verbatim', async () => {
      const user = userEvent.setup()
      mockCalculate.mockRejectedValue(new ApiError('division by zero'))
      render(<Calculator />)

      await submit(user, { operation: 'divide', a: '1', b: '0' })

      expect(await screen.findByText('division by zero')).toBeInTheDocument()
    })

    it('shows the network message when the backend is unreachable', async () => {
      const user = userEvent.setup()
      mockCalculate.mockRejectedValue(
        new ApiError('Could not reach the server. Is the backend running?'),
      )
      render(<Calculator />)

      await submit(user, { operation: 'add', a: '2', b: '3' })

      expect(
        await screen.findByText('Could not reach the server. Is the backend running?'),
      ).toBeInTheDocument()
    })

    it('falls back to a generic message for an unexpected failure', async () => {
      const user = userEvent.setup()
      // Not an ApiError — a bug on this side, whose text should not reach the UI.
      mockCalculate.mockRejectedValue(new TypeError('cannot read properties of null'))
      render(<Calculator />)

      await submit(user, { operation: 'add', a: '2', b: '3' })

      expect(await screen.findByText('Something went wrong.')).toBeInTheDocument()
      expect(screen.queryByText(/cannot read properties/)).not.toBeInTheDocument()
    })

    it('disables the button while the request is in flight', async () => {
      const user = userEvent.setup()
      let resolveCalculation: (value: { result: number }) => void = () => {}
      mockCalculate.mockReturnValue(
        new Promise((resolve) => {
          resolveCalculation = resolve
        }),
      )
      render(<Calculator />)

      await submit(user, { operation: 'add', a: '2', b: '3' })

      expect(submitButton()).toBeDisabled()
      expect(submitButton()).toHaveTextContent('Calculating…')

      resolveCalculation({ result: 5 })

      expect(await screen.findByText('Result: 5')).toBeInTheDocument()
      expect(submitButton()).toBeEnabled()
      expect(submitButton()).toHaveTextContent('Calculate')
    })
  })

  describe('stale output', () => {
    // A result computed from inputs that have since changed is a wrong answer on
    // screen, so any edit clears it.
    it.each([
      { name: 'an operand is edited', edit: async (user: ReturnType<typeof userEvent.setup>) => user.type(operandA(), '7') },
      { name: 'the operation is changed', edit: async (user: ReturnType<typeof userEvent.setup>) => user.selectOptions(operationSelect(), 'multiply') },
    ])('clears the previous result when $name', async ({ edit }) => {
      const user = userEvent.setup()
      mockCalculate.mockResolvedValue({ result: 5 })
      render(<Calculator />)

      await submit(user, { operation: 'add', a: '2', b: '3' })
      expect(await screen.findByText('Result: 5')).toBeInTheDocument()

      await edit(user)

      expect(screen.queryByText('Result: 5')).not.toBeInTheDocument()
    })
  })
})
