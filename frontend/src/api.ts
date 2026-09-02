import {
    ListApplications, GetApplication,
    CreateApplication, UpdateApplication, UpdateApplicationStatus, DeleteApplication,
    CreateInterview, UpdateInterview, DeleteInterview,
    CreateQAItem, UpdateQAItem, DeleteQAItem,
    GetStats, ExportData,
} from '../wailsjs/go/main/App';
import type {Application, ApplicationInput, Interview, InterviewInput, QAItem, QAItemInput, Stats} from './types';

export const api = {
    listApplications: (): Promise<Application[]> => ListApplications() as unknown as Promise<Application[]>,
    getApplication: (id: number): Promise<Application> => GetApplication(id) as unknown as Promise<Application>,
    createApplication: (inp: ApplicationInput): Promise<Application> => CreateApplication(inp) as unknown as Promise<Application>,
    updateApplication: (id: number, inp: ApplicationInput): Promise<Application> => UpdateApplication(id, inp) as unknown as Promise<Application>,
    updateStatus: (id: number, status: string): Promise<Application> => UpdateApplicationStatus(id, status) as unknown as Promise<Application>,
    deleteApplication: (id: number): Promise<void> => DeleteApplication(id),

    createInterview: (inp: InterviewInput): Promise<Interview> => CreateInterview(inp) as unknown as Promise<Interview>,
    updateInterview: (id: number, inp: InterviewInput): Promise<Interview> => UpdateInterview(id, inp) as unknown as Promise<Interview>,
    deleteInterview: (id: number): Promise<void> => DeleteInterview(id),

    createQA: (inp: QAItemInput): Promise<QAItem> => CreateQAItem(inp) as unknown as Promise<QAItem>,
    updateQA: (id: number, inp: QAItemInput): Promise<QAItem> => UpdateQAItem(id, inp) as unknown as Promise<QAItem>,
    deleteQA: (id: number): Promise<void> => DeleteQAItem(id),

    getStats: (): Promise<Stats> => GetStats() as unknown as Promise<Stats>,
    exportData: (): Promise<string> => ExportData(),
};
