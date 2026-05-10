import React from 'react'
import config from '@payload-config'
import type { ServerFunctionClient } from 'payload'
import { RootLayout, handleServerFunctions } from '@payloadcms/next/layouts'
import { importMap } from './cms-admin/importMap.js'

import '@payloadcms/next/css'
import './custom.scss'

const serverFn: ServerFunctionClient = async function (args) {
  'use server'
  return handleServerFunctions({ ...args, config, importMap })
}

export default function PayloadLayout({ children }: { children: React.ReactNode }) {
  return (
    <RootLayout config={config} importMap={importMap} serverFunction={serverFn}>
      {children}
    </RootLayout>
  )
}
