import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { BackupRecord } from '@/api/admin/backup'
import {
  createDefaultBackupS3Config,
  createDefaultBackupScheduleConfig
} from '../backupView'
import BackupOperationsCard from '../backup/BackupOperationsCard.vue'
import BackupR2GuideModal from '../backup/BackupR2GuideModal.vue'
import BackupS3ConfigCard from '../backup/BackupS3ConfigCard.vue'
import BackupScheduleCard from '../backup/BackupScheduleCard.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function createRecord(overrides: Partial<BackupRecord> = {}): BackupRecord {
  return {
    id: 'backup_1',
    status: 'completed',
    backup_type: 'full',
    file_name: 'backup.tar.gz',
    s3_key: 'backups/backup.tar.gz',
    size_bytes: 2048,
    triggered_by: 'manual',
    started_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('backup local components', () => {
  it('renders S3 config card and emits guide, test, and save actions', async () => {
    const wrapper = mount(BackupS3ConfigCard, {
      props: {
        form: createDefaultBackupS3Config(),
        secretConfigured: true,
        saving: false,
        testing: false
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('open-guide')?.length).toBe(1)
    expect(wrapper.emitted('test')?.length).toBe(1)
    expect(wrapper.emitted('save')?.length).toBe(1)
  })

  it('renders schedule card and emits save', async () => {
    const wrapper = mount(BackupScheduleCard, {
      props: {
        form: createDefaultBackupScheduleConfig(),
        saving: false
      }
    })

    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('save')?.length).toBe(1)
  })

  it('renders operations card and emits row actions', async () => {
    const wrapper = mount(BackupOperationsCard, {
      props: {
        backups: [
          createRecord(),
          createRecord({ id: 'backup_2', status: 'running', progress: 'uploading' })
        ],
        loading: false,
        creating: false,
        restoringId: '',
        manualExpireDays: 14
      }
    })

    await wrapper.find('input[type="number"]').setValue('21')

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')
    await buttons[3].trigger('click')
    await buttons[4].trigger('click')

    expect(wrapper.emitted('update:manualExpireDays')?.[0]).toEqual([21])
    expect(wrapper.emitted('create')?.length).toBe(1)
    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('download')?.[0]).toEqual(['backup_1'])
    expect(wrapper.emitted('restore')?.[0]).toEqual(['backup_1'])
    expect(wrapper.emitted('remove')?.[0]).toEqual(['backup_1'])
    expect(wrapper.text()).toContain('admin.backup.progress.uploading')
  })

  it('renders empty state when there are no backups', () => {
    const wrapper = mount(BackupOperationsCard, {
      props: {
        backups: [],
        loading: false,
        creating: false,
        restoringId: '',
        manualExpireDays: 14
      }
    })

    expect(wrapper.text()).toContain('admin.backup.empty')
  })

  it('renders R2 guide modal and emits close', async () => {
    const wrapper = mount(BackupR2GuideModal, {
      props: {
        show: true
      },
      global: {
        stubs: {
          teleport: true,
          transition: false
        }
      }
    })

    expect(wrapper.text()).toContain('admin.backup.r2Guide.title')

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')

    expect(wrapper.emitted('close')?.length).toBe(2)
  })
})
