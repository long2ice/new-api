/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import type { UseFormReturn } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

import {
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'

type Props = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<any>
}

export function BepusdtSettingsSection({ form }: Props) {
  const { t } = useTranslation()

  return (
    <div className='space-y-4'>
      <div>
        <h3 className='text-lg font-medium'>{t('BEPUSDT Gateway')}</h3>
        <p className='text-muted-foreground text-sm'>
          {t('Configuration for BEPUSDT (USDT) payment integration')}
        </p>
      </div>

      <div className='rounded-md bg-teal-50 p-4 text-sm text-teal-900 dark:bg-teal-950 dark:text-teal-100'>
        <p className='mb-2 font-medium'>{t('Webhook Configuration:')}</p>
        <ul className='list-inside list-disc space-y-1'>
          <li>
            {t('Notify URL:')}{' '}
            <code className='rounded bg-teal-100 px-1 py-0.5 text-xs dark:bg-teal-900'>
              {'{ServerAddress}/api/bepusdt/notify'}
            </code>
          </li>
          <li>
            {t(
              'Leave Trade Type empty to use create-order; set it to use create-transaction with trade_type.'
            )}
          </li>
        </ul>
      </div>

      <FormField
        control={form.control}
        name='BepusdtEnabled'
        render={({ field }) => (
          <SettingsSwitchItem>
            <SettingsSwitchContent>
              <FormLabel>{t('Enable BEPUSDT')}</FormLabel>
              <FormDescription>
                {t(
                  'Show BEPUSDT as a top-up payment method when configured'
                )}
              </FormDescription>
            </SettingsSwitchContent>
            <FormControl>
              <Switch
                checked={field.value}
                onCheckedChange={field.onChange}
              />
            </FormControl>
          </SettingsSwitchItem>
        )}
      />

      <div className='grid gap-6 md:grid-cols-2'>
        <FormField
          control={form.control}
          name='BepusdtUrl'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Gateway URL')}</FormLabel>
              <FormControl>
                <Input
                  placeholder='https://pay.example.com'
                  autoComplete='off'
                  {...field}
                />
              </FormControl>
              <FormDescription>
                {t('Base URL of the BEPUSDT gateway, without path')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='BepusdtApiKey'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('API Key')}</FormLabel>
              <FormControl>
                <Input
                  type='password'
                  placeholder={t('Enter new key to update')}
                  autoComplete='new-password'
                  {...field}
                />
              </FormControl>
              <FormDescription>
                {t('Used for request and callback MD5 signatures')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className='grid gap-6 md:grid-cols-2'>
        <FormField
          control={form.control}
          name='BepusdtFiat'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Fiat currency')}</FormLabel>
              <FormControl>
                <Input placeholder='USD' autoComplete='off' {...field} />
              </FormControl>
              <FormDescription>
                {t('Defaults to USD when empty')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='BepusdtCurrencies'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Crypto currencies (optional)')}</FormLabel>
              <FormControl>
                <Input
                  placeholder='USDT.TRC20,USDT.BEP20'
                  autoComplete='off'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className='grid gap-6 md:grid-cols-2'>
        <FormField
          control={form.control}
          name='BepusdtTradeType'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Trade type (optional)')}</FormLabel>
              <FormControl>
                <Input placeholder='usdt' autoComplete='off' {...field} />
              </FormControl>
              <FormDescription>
                {t('When set, uses create-transaction with this trade_type')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='BepusdtUnitPrice'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Unit price')}</FormLabel>
              <FormControl>
                <Input
                  type='number'
                  min={0}
                  step='0.01'
                  value={field.value}
                  onChange={(event) =>
                    field.onChange(Number(event.target.value))
                  }
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className='grid gap-6 md:grid-cols-2'>
        <FormField
          control={form.control}
          name='BepusdtMinTopUp'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Minimum top-up')}</FormLabel>
              <FormControl>
                <Input
                  type='number'
                  min={1}
                  step='1'
                  value={field.value}
                  onChange={(event) =>
                    field.onChange(Number(event.target.value))
                  }
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='BepusdtNotifyUrl'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Notify URL override (optional)')}</FormLabel>
              <FormControl>
                <Input
                  placeholder='https://api.example.com/api/bepusdt/notify'
                  autoComplete='off'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <FormField
        control={form.control}
        name='BepusdtReturnUrl'
        render={({ field }) => (
          <FormItem>
            <FormLabel>{t('Return URL override (optional)')}</FormLabel>
            <FormControl>
              <Input
                placeholder='https://example.com/console/topup'
                autoComplete='off'
                {...field}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  )
}
