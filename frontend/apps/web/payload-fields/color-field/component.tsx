'use client'

import { useField } from '@payloadcms/ui'
import { HexColorInput, HexColorPicker } from 'react-colorful'
import type { TextFieldClientProps } from 'payload'
import { useCallback, useState } from 'react'

export const ColorField: React.FC<TextFieldClientProps> = ({ field, path }) => {
  const { value, setValue } = useField<string>({ path: path ?? field.name })
  const [open, setOpen] = useState(false)
  const [color, setColor] = useState<string>((value as string) || '#6A35E0')

  const handleChange = useCallback(
    (newColor: string) => {
      setColor(newColor)
      setValue(newColor)
    },
    [setValue],
  )
  const label = field.label || field.name

  return (
    <div className="field-type text">
      <label className="field-label">{typeof label === 'string' ? label : field.name}</label>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center', position: 'relative' }}>
        <button
          type="button"
          onClick={() => setOpen(!open)}
          aria-label="Open color picker"
          style={{
            width: 32,
            height: 32,
            borderRadius: 6,
            border: '1px solid #ccc',
            background: color,
            cursor: 'pointer',
            flexShrink: 0,
          }}
        />
        <HexColorInput
          color={color}
          onChange={handleChange}
          prefixed
          style={{
            flex: 1,
            padding: '8px 10px',
            border: '1px solid #ccc',
            borderRadius: 6,
            fontFamily: 'monospace',
            fontSize: 14,
          }}
        />
      </div>
      {open && (
        <div style={{ marginTop: 8, position: 'relative', zIndex: 10 }}>
          <HexColorPicker color={color} onChange={handleChange} />
        </div>
      )}
      {field.admin?.description && (
        <p className="field-description" style={{ marginTop: 4, fontSize: 12, opacity: 0.7 }}>
          {typeof field.admin.description === 'string' ? field.admin.description : ''}
        </p>
      )}
    </div>
  )
}

export default ColorField
